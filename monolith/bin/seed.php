<?php
declare(strict_types=1);

/**
 * Сидер стенда: данные, на которых видно списки, фильтры и отчёты.
 *
 * 3 агентства, 10 сотрудников, 2 проекта, 200 лотов, 300 фиксаций
 * с разными статусами и разбросом по датам.
 *
 * ВАЖНО про подключение. Фиксации лежат в схеме integration, которой
 * владеет Go, и у app_user там только select. Поэтому сидер ходит под
 * суперпользователем — он намеренно нарушает границу контуров, потому
 * что это стенд, а не продакшен. В настоящем проекте демо-данные
 * интеграционного контура наливал бы отдельный сидер на стороне Go.
 *
 * Запуск:  make seed
 */

require __DIR__ . '/../vendor/autoload.php';

use Broker\Monolith\Log;

$dsn      = getenv('SEED_DB_DSN') ?: (getenv('APP_DB_DSN') ?: 'pgsql:host=127.0.0.1;port=5432;dbname=broker');
$user     = getenv('SEED_DB_USER') ?: 'postgres';
$password = getenv('SEED_DB_PASSWORD') ?: 'postgres';

$pdo = new PDO($dsn, $user, $password, [
    PDO::ATTR_ERRMODE            => PDO::ERRMODE_EXCEPTION,
    PDO::ATTR_DEFAULT_FETCH_MODE => PDO::FETCH_ASSOC,
    PDO::ATTR_EMULATE_PREPARES   => false,
]);

/** Детерминированный uuid v4 из mt_rand: сид фиксирован, данные воспроизводимы. */
function uuid(): string
{
    $bytes    = random_bytes(16);
    $bytes[6] = chr((ord($bytes[6]) & 0x0f) | 0x40);
    $bytes[8] = chr((ord($bytes[8]) & 0x3f) | 0x80);

    return vsprintf('%s%s-%s-%s-%s-%s%s%s', str_split(bin2hex($bytes), 4));
}

// Всё одной транзакцией: наполовину налитый стенд хуже пустого —
// он выглядит рабочим и врёт.
$pdo->beginTransaction();

try {
    // Идемпотентность сидера: гоняем его многократно, каждый раз с нуля.
    // Порядок обратный порядку зависимостей — сначала то, на что ссылаются.
    $pdo->exec('delete from integration.fixations');
    $pdo->exec('delete from app.refresh_sessions');
    $pdo->exec('delete from app.lots');
    $pdo->exec('update app.users set agency_id = null');
    $pdo->exec('delete from app.users');
    $pdo->exec('delete from app.projects');
    $pdo->exec('delete from app.agencies');

    // ── Агентства ────────────────────────────────────────────────────
    // Разные статусы намеренно: на списках должно быть видно, что
    // заблокированному агентству фиксировать нельзя.
    $agencies = [
        ['id' => uuid(), 'name' => 'АН «Первый метр»',   'inn' => '7701234567', 'status' => 'active'],
        ['id' => uuid(), 'name' => 'АН «Ключи города»',  'inn' => '7702345678', 'status' => 'active'],
        ['id' => uuid(), 'name' => 'АН «Новосёл»',       'inn' => '7703456789', 'status' => 'blocked'],
    ];

    $stmt = $pdo->prepare(
        'insert into app.agencies (id, name, inn, status) values (:id, :name, :inn, :status)'
    );
    foreach ($agencies as $agency) {
        $stmt->execute($agency);
    }

    // ── Сотрудники ───────────────────────────────────────────────────
    // 8 брокеров по агентствам + 2 сотрудника застройщика без агентства.
    // Пароль у всех один: bcrypt от 'password'. Стенд, не прод.
    $passwordHash = password_hash('password', PASSWORD_BCRYPT);

    $users = [];
    $roles = ['agency_owner', 'broker_team_lead', 'broker_team_member', 'broker_team_member'];

    for ($i = 0; $i < 8; $i++) {
        $agency  = $agencies[$i % 2];              // только в два активных
        $users[] = [
            'id'            => uuid(),
            'agency_id'     => $agency['id'],
            'email'         => sprintf('broker%d@%s.test', $i + 1, $i % 2 === 0 ? 'perviy-metr' : 'kluchi'),
            'user_role'     => $roles[$i % count($roles)],
            'password_hash' => $passwordHash,
            'last_name'     => 'Брокеров',
            'first_name'    => 'Брокер' . ($i + 1),
        ];
    }

    foreach (['account_manager', 'sales_manager'] as $index => $role) {
        $users[] = [
            'id'            => uuid(),
            'agency_id'     => null,               // сотрудник застройщика
            'email'         => sprintf('%s@developer.test', $role),
            'user_role'     => $role,
            'password_hash' => $passwordHash,
            'last_name'     => 'Застройщиков',
            'first_name'    => 'Сотрудник' . ($index + 1),
        ];
    }

    $stmt = $pdo->prepare(
        'insert into app.users (id, agency_id, email, user_role, password_hash, last_name, first_name)
         values (:id, :agency_id, :email, :user_role, :password_hash, :last_name, :first_name)'
    );
    foreach ($users as $user) {
        $stmt->execute($user);
    }

    // ── Проекты ──────────────────────────────────────────────────────
    $projects = [
        ['id' => uuid(), 'name' => 'ЖК «Северная звезда»', 'status' => 'active'],
        ['id' => uuid(), 'name' => 'ЖК «Речной парк»',     'status' => 'active'],
    ];

    $stmt = $pdo->prepare('insert into app.projects (id, name, status) values (:id, :name, :status)');
    foreach ($projects as $project) {
        $stmt->execute($project);
    }

    // ── Лоты ─────────────────────────────────────────────────────────
    // 200 штук, по 100 на проект, с разбросом по корпусам, этажам,
    // комнатности и статусам — иначе фильтры витрины нечем проверять.
    $stmt = $pdo->prepare(
        'insert into app.lots
            (id, project_id, external_id, building, number, floor, rooms, area_m2, price_kopecks, status, synced_at)
         values
            (:id, :project_id, :external_id, :building, :number, :floor, :rooms, :area_m2, :price_kopecks, :status, now())'
    );

    $lotStatuses = ['free', 'free', 'free', 'booked', 'sold'];  // перекос в свободные — как в жизни

    for ($p = 0; $p < count($projects); $p++) {
        for ($i = 1; $i <= 100; $i++) {
            $floor = (int)ceil($i / 4);
            $rooms = (($i - 1) % 4) + 1;

            $stmt->execute([
                'id'            => uuid(),
                'project_id'    => $projects[$p]['id'],
                'external_id'   => sprintf('PB-%d-%03d', $p + 1, $i),
                'building'      => 'Корпус ' . (($i <= 50) ? '1' : '2'),
                'number'        => (string)(100 + $i),
                'floor'         => $floor,
                'rooms'         => $rooms,
                'area_m2'       => 28.5 + $rooms * 12.4,
                'price_kopecks' => (5_200_000 + $rooms * 1_800_000 + $floor * 40_000) * 100,
                'status'        => $lotStatuses[($i + $p) % count($lotStatuses)],
            ]);
        }
    }

    // ── Фиксации ─────────────────────────────────────────────────────
    // 300 штук за последние полгода. Статусы и даты разные: на отчёте
    // «сколько фиксаций дошло до сделки» должна быть видна воронка,
    // а не один столбик.
    //
    // Активная фиксация в проекте может быть только одна на телефон —
    // это держит частичный уникальный индекс. Поэтому телефон у каждой
    // фиксации свой, и на индекс сидер не наступает.
    $stmt = $pdo->prepare(
        'insert into integration.fixations
            (id, fixed_at, expires_at, status, agency_id, fix_by, fix_for, project_id, phone_hash)
         values
            (:id, :fixed_at, :expires_at, :status, :agency_id, :fix_by, :fix_for, :project_id, :phone_hash)'
    );

    $brokers = array_values(array_filter($users, static fn(array $u): bool => $u['agency_id'] !== null));
    // Перекос в active: так выглядит живая база в разгар продаж.
    $fixationStatuses = ['active', 'active', 'active', 'converted', 'expired', 'removed'];

    for ($i = 0; $i < 300; $i++) {
        $broker = $brokers[$i % count($brokers)];
        $status = $fixationStatuses[$i % count($fixationStatuses)];

        // Разброс на 180 дней назад.
        $daysAgo   = $i % 180;
        $fixedAt   = (new DateTimeImmutable("-{$daysAgo} days"))->format(DATE_ATOM);
        // Протухшие — с датой окончания в прошлом, остальные — в будущем.
        $expiresAt = $status === 'expired'
            ? (new DateTimeImmutable('-' . max(1, $daysAgo - 60) . ' days'))->format(DATE_ATOM)
            : (new DateTimeImmutable("-{$daysAgo} days +60 days"))->format(DATE_ATOM);

        $stmt->execute([
            'id'         => uuid(),
            'fixed_at'   => $fixedAt,
            'expires_at' => $expiresAt,
            'status'     => $status,
            'agency_id'  => $broker['agency_id'],
            'fix_by'     => $broker['id'],
            'fix_for'    => $broker['id'],
            'project_id' => $projects[$i % count($projects)]['id'],
            // Хэш, а не телефон: в integration.fixations телефона в открытом
            // виде нет и не должно быть. Здесь просто правдоподобная строка —
            // настоящий хэш считает fixationservice своей солью.
            'phone_hash' => base64_encode(hash('sha256', sprintf('+79%09d', 100000000 + $i), true)),
        ]);
    }

    $pdo->commit();

    Log::line(sprintf(
        'seed done: agencies=%d users=%d projects=%d lots=%d fixations=%d',
        count($agencies),
        count($users),
        count($projects),
        200,
        300
    ));

    echo "Агентства:\n";
    foreach ($agencies as $agency) {
        echo "  {$agency['id']}  {$agency['status']}  {$agency['name']}\n";
    }
    echo "Проекты:\n";
    foreach ($projects as $project) {
        echo "  {$project['id']}  {$project['name']}\n";
    }
    echo "Логин любого брокера: broker1@perviy-metr.test / password\n";
} catch (\Throwable $e) {
    $pdo->rollBack();
    Log::line('seed failed: ' . $e->getMessage());
    exit(1);
}
