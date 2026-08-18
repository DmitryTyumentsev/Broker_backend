<?php
declare(strict_types=1);

/**
 * Монолит-заглушка. Чистый PHP 8.0, без Laravel.
 *
 * Это НЕ пример того, как надо писать монолит. Это минимум, который
 * нужен Go-контуру, чтобы с той стороны границы кто-то отвечал:
 * справочник агентств, шахматка, приём уведомлений и health.
 * Что из этого в настоящем Laravel делал бы фреймворк — в monolith/README.md.
 *
 * Роутер здесь примитивный: match по методу и пути. Регулярка на {id}
 * ровно одна. Как только их станет три, это надо будет заменить на
 * настоящий роутер — но к тому моменту у нас уже будет Laravel.
 */

require __DIR__ . '/../vendor/autoload.php';

use Broker\Monolith\Bus;
use Broker\Monolith\Db;
use Broker\Monolith\Log;

$method = $_SERVER['REQUEST_METHOD'] ?? 'GET';
$path   = parse_url($_SERVER['REQUEST_URI'] ?? '/', PHP_URL_PATH) ?: '/';

header('Content-Type: application/json; charset=utf-8');

try {
    // ── GET /internal/v1/health ──────────────────────────────────────
    // Проверяет не «процесс жив», а «процесс может работать»: живой PHP
    // с мёртвой базой бесполезен, и балансировщик должен об этом узнать.
    if ($method === 'GET' && $path === '/internal/v1/health') {
        $dbOk = true;
        try {
            Db::pdo()->query('select 1');
        } catch (\Throwable $e) {
            $dbOk = false;
        }

        respond($dbOk ? 200 : 503, [
            'status' => $dbOk ? 'ok' : 'degraded',
            'db'     => $dbOk,
        ]);
    }

    // ── GET /internal/v1/agencies/{id} ───────────────────────────────
    // Go спрашивает: «что это за агентство и можно ли ему фиксировать».
    // Отдаём статус, а не булев флаг: решение «пускать или нет» принимает
    // вызывающий, наше дело — состояние.
    if ($method === 'GET' && preg_match('#^/internal/v1/agencies/([0-9a-fA-F-]{36})$#', $path, $m) === 1) {
        $stmt = Db::pdo()->prepare(
            'select a.id, a.name, a.inn, a.status, a.created_at,
                    (select count(*) from app.users u where u.agency_id = a.id) as employees_count
               from app.agencies a
              where a.id = :id'
        );
        $stmt->execute(['id' => $m[1]]);
        $agency = $stmt->fetch();

        if ($agency === false) {
            respond(404, ['error' => 'agency not found']);
        }

        $agency['employees_count'] = (int)$agency['employees_count'];
        respond(200, $agency);
    }

    // ── GET /internal/v1/lots?project_id=... ─────────────────────────
    // Шахматка. Лимит стоит всегда: ручка внутренняя, но «отдай все лоты»
    // на живом корпусе — это десятки тысяч строк и таймаут у вызывающего.
    if ($method === 'GET' && $path === '/internal/v1/lots') {
        $projectId = $_GET['project_id'] ?? '';
        if (preg_match('#^[0-9a-fA-F-]{36}$#', (string)$projectId) !== 1) {
            respond(400, ['error' => 'project_id is required and must be uuid']);
        }

        $limit = (int)($_GET['limit'] ?? 100);
        $limit = max(1, min($limit, 500));

        $stmt = Db::pdo()->prepare(
            'select id, project_id, building, number, floor, rooms, area_m2,
                    price_kopecks, status, synced_at
               from app.lots
              where project_id = :project_id
              order by building, number
              limit :limit'
        );
        $stmt->bindValue('project_id', $projectId);
        $stmt->bindValue('limit', $limit, PDO::PARAM_INT);
        $stmt->execute();

        respond(200, ['items' => $stmt->fetchAll(), 'limit' => $limit]);
    }

    // ── POST /internal/v1/notifications ──────────────────────────────
    // Go просит монолит уведомить человека: письма, пуши и шаблоны живут
    // в монолите, и дублировать их в Go незачем.
    //
    // Отвечаем 202, а не 200: мы приняли задание, но ещё не отправили.
    // Отправка асинхронная, и врать про результат нельзя.
    if ($method === 'POST' && $path === '/internal/v1/notifications') {
        $raw  = file_get_contents('php://input') ?: '';
        $body = json_decode($raw, true);

        if (!is_array($body) || !isset($body['type'], $body['recipient_id'])) {
            respond(400, ['error' => 'type and recipient_id are required']);
        }

        Log::line(sprintf(
            'notification accepted: type=%s recipient=%s',
            (string)$body['type'],
            (string)$body['recipient_id']
        ));

        respond(202, ['accepted' => true]);
    }

    // ── POST /internal/v1/agencies/{id}/approve ──────────────────────
    // Не в списке обязательных ручек, но без неё нечем проверить
    // публикацию в admin.events. Админка застройщика одобряет агентство —
    // Go должен об этом узнать.
    if ($method === 'POST' && preg_match('#^/internal/v1/agencies/([0-9a-fA-F-]{36})/approve$#', $path, $m) === 1) {
        $stmt = Db::pdo()->prepare(
            "update app.agencies set status = 'active', updated_at = now()
              where id = :id and status <> 'active'
          returning id, name, status"
        );
        $stmt->execute(['id' => $m[1]]);
        $agency = $stmt->fetch();

        if ($agency === false) {
            respond(409, ['error' => 'agency not found or already active']);
        }

        // Публикуем ПОСЛЕ успешного update. Порядок важен: сначала факт
        // в базе, потом сообщение о факте. Наоборот — и подписчик узнает
        // о том, чего не произошло.
        Bus::publish(Bus::EXCHANGE_ADMIN, 'agency.approved', [
            'agency_id'   => $agency['id'],
            'name'        => $agency['name'],
            'occurred_at' => gmdate('c'),
        ]);

        respond(200, $agency);
    }

    // ── POST /internal/v1/agencies/{id}/tariff ───────────────────────
    // Смена тарифа агентства — второе событие в admin.events.
    if ($method === 'POST' && preg_match('#^/internal/v1/agencies/([0-9a-fA-F-]{36})/tariff$#', $path, $m) === 1) {
        $raw  = file_get_contents('php://input') ?: '';
        $body = json_decode($raw, true);

        if (!is_array($body) || !isset($body['commission_percent'])) {
            respond(400, ['error' => 'commission_percent is required']);
        }

        Bus::publish(Bus::EXCHANGE_ADMIN, 'tariff.changed', [
            'agency_id'          => $m[1],
            'commission_percent' => (float)$body['commission_percent'],
            'occurred_at'        => gmdate('c'),
        ]);

        respond(202, ['accepted' => true]);
    }

    respond(404, ['error' => 'route not found']);
} catch (\Throwable $e) {
    // Наружу — без подробностей: текст исключения содержит имена таблиц
    // и куски запроса. В лог — целиком.
    Log::line('unhandled: ' . $e->getMessage());
    respond(500, ['error' => 'internal error']);
}

/**
 * Пишет ответ и завершает запрос.
 *
 * Возвращаемого типа нет намеренно: подошёл бы `never`, но он появился
 * только в PHP 8.1, а мы на 8.0 — как и прод.
 *
 * @param array<string,mixed>|list<mixed> $payload
 */
function respond(int $status, array $payload): void
{
    http_response_code($status);
    echo json_encode($payload, JSON_UNESCAPED_UNICODE | JSON_PRETTY_PRINT);
    exit;
}
