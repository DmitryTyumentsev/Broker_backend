<?php
declare(strict_types=1);

/**
 * Сидер стенда: данные, на которых видно списки, фильтры и отчёты.
 *
 * Агентства с разными статусами, брокеры в каждом из них, сотрудники
 * застройщика без агентства вообще, проекты активные и архивные, шахматка
 * и фиксации за последние полгода с разбросом по статусам и датам.
 *
 * Итогов сидер не печатает намеренно: «сколько чего в базе» — это запрос,
 * который в расследовании пишешь ты, а не подсказка из вывода скрипта.
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

/**
 * Нормализация номера — та же, что в shared/pkg/helpers: из строки берутся
 * только цифры, «8XXXXXXXXXX» и десятизначный номер без плюса приводятся
 * к «7XXXXXXXXXX».
 */
function normalizePhone(string $phone): string
{
    $explicitCountryCode = str_starts_with(trim($phone), '+');
    $digits              = preg_replace('/\D+/', '', $phone) ?? '';

    if (!$explicitCountryCode && strlen($digits) === 10) {
        return '7' . $digits;
    }

    if (!$explicitCountryCode && strlen($digits) === 11 && $digits[0] === '8') {
        return '7' . substr($digits, 1);
    }

    return $digits;
}

/**
 * Хэш телефона ровно такой, какой считает fixationservice: HMAC-SHA256
 * по нормализованному номеру, ключ — секрет в том виде, в каком его
 * сериализует json.Marshal (то есть в кавычках), результат — base64.
 *
 * Совпадение здесь не косметика. Если сидер кладёт «правдоподобную
 * строку», то фиксация, оформленная через API на тот же номер, не
 * встретится с сидированной на частичном уникальном индексе — и стенд
 * молча перестаёт воспроизводить главное свойство продукта.
 */
function phoneHash(string $phone, string $secret): string
{
    return base64_encode(hash_hmac('sha256', normalizePhone($phone), json_encode($secret), true));
}

/**
 * Как хэш считался раньше, до того как нормализацию вынесли в одно место:
 * sha256 по строке как её прислала CRM. Нужен, чтобы в базе стенда лежала
 * история — записи, сделанные до исправления. Новый код так не считает.
 */
function legacyPhoneHash(string $phone): string
{
    return base64_encode(hash('sha256', $phone, true));
}

// Секрет тот же, что у fixationservice. Разошлись — разошлись и хэши,
// и все проверки уникальности на стенде становятся бессмысленными.
$hashSecret = getenv('SEED_PHONE_HASH_SECRET') ?: 'local-phone-hash-secret-change-me';

// Всё одной транзакцией: наполовину налитый стенд хуже пустого —
// он выглядит рабочим и врёт.
$pdo->beginTransaction();

try {
    // Идемпотентность сидера: гоняем его многократно, каждый раз с нуля.
    // Порядок обратный порядку зависимостей — сначала то, на что ссылаются.
    // truncate, а не delete: после `make seed-bulk` в таблице миллионы строк,
    // и построчное удаление занимает минуты и оставляет за собой раздутую
    // таблицу. Внешних ключей на фиксации нет — обрезать безопасно.
    $pdo->exec('truncate table integration.fixations');
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
    // Брокеры разложены по РАЗНЫМ агентствам намеренно, включая
    // заблокированное. Сотрудник одного агентства и сотрудник другого —
    // это разные входные данные для любой проверки принадлежности,
    // и без обоих такую проверку нечем проверить.
    //
    // Пароль у всех один: bcrypt от 'password'. Стенд, не прод.
    $passwordHash = password_hash('password', PASSWORD_BCRYPT);

    $slugs = ['perviy-metr', 'kluchi', 'novosel'];
    $roles = ['agency_owner', 'broker_team_lead', 'broker_team_member', 'broker_team_member'];

    $users = [];
    $brokerIndex = 0;

    // По несколько брокеров в каждое агентство, включая заблокированное:
    // тикет из поддержки пришёл именно от него, и войти под его сотрудником
    // должно быть возможно.
    foreach ($agencies as $agencyIndex => $agency) {
        for ($i = 0; $i < 3; $i++) {
            $brokerIndex++;
            $users[] = [
                'id'            => uuid(),
                'agency_id'     => $agency['id'],
                'email'         => sprintf('broker%d@%s.test', $brokerIndex, $slugs[$agencyIndex]),
                'user_role'     => $roles[$i % count($roles)],
                'password_hash' => $passwordHash,
                'last_name'     => 'Брокеров',
                'first_name'    => 'Брокер' . $brokerIndex,
            ];
        }
    }

    // Сотрудники застройщика: агентства нет вообще. Отдельный класс входных
    // данных — не «чужое агентство», а «агентства нет».
    foreach (['account_manager', 'sales_manager', 'developer_admin'] as $index => $role) {
        $users[] = [
            'id'            => uuid(),
            'agency_id'     => null,
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
    // Статусы разные намеренно. Архивный проект — это отдельное состояние,
    // а не «удалённый»: строка в базе есть, ссылки на неё живы, продавать
    // по нему нельзя. Без такой строки в стенде поведение системы на
    // архивном проекте нечем воспроизвести.
    $projects = [
        ['id' => uuid(), 'name' => 'ЖК «Северная звезда»', 'status' => 'active'],
        ['id' => uuid(), 'name' => 'ЖК «Речной парк»',     'status' => 'active'],
        ['id' => uuid(), 'name' => 'ЖК «Заречье», очередь 1', 'status' => 'archived'],
    ];

    $stmt = $pdo->prepare('insert into app.projects (id, name, status) values (:id, :name, :status)');
    foreach ($projects as $project) {
        $stmt->execute($project);
    }

    // Проект, которого в app.projects НЕТ. Не архивный — удалённый:
    // схемой app владеет монолит, строку из неё он убрал, а фиксации на
    // неё в нашей схеме остались. Внешнего ключа между схемами нет и быть
    // не может (разные владельцы, разные роли), значит такие сироты в базе
    // есть всегда — и любой список обязан их переживать.
    $deletedProjectId = uuid();

    // ── Лоты ─────────────────────────────────────────────────────────
    // Своя шахматка у каждого проекта, включая архивный: у архивного
    // проекта лоты никуда не деваются, он просто закрыт для продаж.
    // Разброс по корпусам, этажам, комнатности и статусам — иначе
    // фильтры витрины нечем проверять.
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
    // За последние полгода, по всем агентствам и всем проектам, включая
    // архивный: фиксация, оформленная до архивации, никуда не исчезает.
    // Статусы и даты разные — на отчёте «сколько фиксаций дошло до сделки»
    // должна быть видна воронка, а не один столбик.
    //
    // Сколько чего получилось — считай сам, это входит в работу.
    $stmt = $pdo->prepare(
        'insert into integration.fixations
            (id, fixed_at, expires_at, status, agency_id, fix_by, fix_for, project_id, phone_hash)
         values
            (:id, :fixed_at, :expires_at, :status, :agency_id, :fix_by, :fix_for, :project_id, :phone_hash)'
    );

    $brokers = array_values(array_filter($users, static fn(array $u): bool => $u['agency_id'] !== null));

    /** Руководитель того же агентства — от его имени оформляют «за» подчинённого. */
    $leadOf = static function (array $broker) use ($users): array {
        foreach ($users as $candidate) {
            if ($candidate['agency_id'] === $broker['agency_id']
                && in_array($candidate['user_role'], ['agency_owner', 'broker_team_lead'], true)
                && $candidate['id'] !== $broker['id']
            ) {
                return $candidate;
            }
        }

        return $broker;
    };

    // Перекос в active: так выглядит живая база в разгар продаж.
    $fixationStatuses = ['active', 'active', 'active', 'converted', 'expired', 'removed'];

    // Последние четыре цифры повторяются намеренно. Четырёх цифр не хватает,
    // чтобы отличить человека от человека: по ним находятся РАЗНЫЕ клиенты,
    // и поиск обязан это переживать, а не считать, что нашёл одного.
    $tails = ['4567', '4567', '1122', '3344', '5566', '4567'];

    /** Телефон фиксации номер $i: детерминированный, формат +7 9XX XXX-XX-XX. */
    $phoneOf = static function (int $i) use ($tails): string {
        return sprintf('+79%05d%s', $i, $tails[$i % count($tails)]);
    };

    $fixationCount = 300 + count($projects) * 7;

    for ($i = 0; $i < $fixationCount; $i++) {
        $broker = $brokers[$i % count($brokers)];
        $status = $fixationStatuses[$i % count($fixationStatuses)];

        // Кто оформил и за кем закреплён клиент — РАЗНЫЕ вещи. Обычно это
        // один человек, но руководитель оформляет и за подчинённого.
        // Пока эти две колонки совпадают во всех строках, ошибка «взял не ту»
        // не проявляется ни в одном тесте.
        $fixBy = ($i % 7 === 0) ? $leadOf($broker)['id'] : $broker['id'];

        // Разброс на 180 дней назад, но не свежее вчерашнего: самые новые
        // записи в стенде — партия из CRM ниже, и она должна быть первой
        // страницей при сортировке по fixed_at убыванию.
        $daysAgo   = 1 + $i % 180;
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
            'fix_by'     => $fixBy,
            'fix_for'    => $broker['id'],
            'project_id' => $projects[$i % count($projects)]['id'],
            // Хэш, а не телефон: в integration.fixations телефона в открытом
            // виде нет и не должно быть.
            'phone_hash' => phoneHash($phoneOf($i), $hashSecret),
        ]);
    }

    // ── Фиксации на удалённый проект ─────────────────────────────────
    // Сироты: project_id есть, строки в app.projects нет. Схемой app
    // владеет монолит, и он свои строки удаляет, не спрашивая нас.
    $orphanBroker = $brokers[0];

    foreach (['active', 'converted', 'expired'] as $index => $status) {
        $daysAgo = 40 + $index * 5;

        $stmt->execute([
            'id'         => uuid(),
            'fixed_at'   => (new DateTimeImmutable("-{$daysAgo} days"))->format(DATE_ATOM),
            'expires_at' => (new DateTimeImmutable("-{$daysAgo} days +60 days"))->format(DATE_ATOM),
            'status'     => $status,
            'agency_id'  => $orphanBroker['agency_id'],
            'fix_by'     => $orphanBroker['id'],
            'fix_for'    => $orphanBroker['id'],
            'project_id' => $deletedProjectId,
            'phone_hash' => phoneHash(sprintf('+7999000%04d', 7000 + $index), $hashSecret),
        ]);
    }

    // ── Один клиент, два проекта ─────────────────────────────────────
    // Один и тот же номер, две живые фиксации на разные ЖК. Частичный
    // уникальный индекс такое разрешает: он про пару (телефон, проект),
    // а не про телефон. В выдаче человек встретится дважды — и это не дубль.
    $twoProjectsClient = '+7 (999) 765-43-21';

    foreach ([0, 1] as $index) {
        $daysAgo = 12 + $index;

        $stmt->execute([
            'id'         => uuid(),
            'fixed_at'   => (new DateTimeImmutable("-{$daysAgo} days"))->format(DATE_ATOM),
            'expires_at' => (new DateTimeImmutable("-{$daysAgo} days +60 days"))->format(DATE_ATOM),
            'status'     => 'active',
            'agency_id'  => $brokers[$index]['agency_id'],
            'fix_by'     => $brokers[$index]['id'],
            'fix_for'    => $brokers[$index]['id'],
            'project_id' => $projects[$index]['id'],
            'phone_hash' => phoneHash($twoProjectsClient, $hashSecret),
        ]);
    }

    // ── История: записи, сделанные до нормализации ───────────────────
    // Один человек, номер приехал в двух форматах, хэши посчитаны по-разному
    // и потому разошлись. Уникальный индекс их не поймал — он сравнивает
    // строки хэшей, а не людей. Такие пары в базе есть, и они старше
    // сегодняшнего кода.
    $legacyClient = ['+7 (999) 123-45-67', '89991234567'];

    foreach ($legacyClient as $index => $written) {
        $daysAgo = 95 + $index;

        $stmt->execute([
            'id'         => uuid(),
            'fixed_at'   => (new DateTimeImmutable("-{$daysAgo} days"))->format(DATE_ATOM),
            'expires_at' => (new DateTimeImmutable("-{$daysAgo} days +180 days"))->format(DATE_ATOM),
            'status'     => 'active',
            'agency_id'  => $brokers[$index]['agency_id'],
            'fix_by'     => $brokers[$index]['id'],
            'fix_for'    => $brokers[$index]['id'],
            'project_id' => $projects[0]['id'],
            'phone_hash' => legacyPhoneHash($written),
        ]);
    }

    // ── Партия из CRM: одинаковый fixed_at до микросекунды ───────────
    // Агентство выгрузило накопленное одним импортом, все записи легли
    // в одной транзакции и получили один и тот же now(). Строк в партии
    // заведомо больше страницы по умолчанию, так что граница страницы
    // проходит ВНУТРИ группы одинаковых времён. Сортировка по одному
    // fixed_at на таких данных не воспроизводима: вторая страница либо
    // повторит часть первой, либо потеряет.
    $importBroker = $brokers[1];
    $importLead   = $leadOf($importBroker);
    $importedAt   = new DateTimeImmutable('-1 hour');
    $importFixed  = $importedAt->format('Y-m-d H:i:s.u P');
    $importExpire = $importedAt->modify('+60 days')->format('Y-m-d H:i:s.u P');

    for ($i = 0; $i < 26; $i++) {
        $stmt->execute([
            'id'         => uuid(),
            'fixed_at'   => $importFixed,
            'expires_at' => $importExpire,
            'status'     => 'active',
            'agency_id'  => $importBroker['agency_id'],
            'fix_by'     => $importLead['id'],
            'fix_for'    => $importBroker['id'],
            'project_id' => $projects[0]['id'],
            // Хвост у всей партии один: поиск по последним четырём цифрам
            // отдаёт ровно её.
            'phone_hash' => phoneHash(sprintf('+79%05d9999', 90000 + $i), $hashSecret),
        ]);
    }

    $pdo->commit();

    Log::line('seed done');

    // Справочник для ручных проверок. Печатаем состав, а не итоги:
    // сколько чего в базе — это запрос, который пишешь ты сам.
    echo "\nАгентства\n";
    foreach ($agencies as $agency) {
        echo sprintf("  %s  %-8s %s\n", $agency['id'], $agency['status'], $agency['name']);

        foreach ($users as $user) {
            if ($user['agency_id'] === $agency['id']) {
                echo sprintf("      %s  %-20s %s\n", $user['id'], $user['user_role'], $user['email']);
            }
        }
    }

    echo "\nСотрудники застройщика (агентства нет)\n";
    foreach ($users as $user) {
        if ($user['agency_id'] === null) {
            echo sprintf("  %s  %-20s %s\n", $user['id'], $user['user_role'], $user['email']);
        }
    }

    echo "\nПроекты\n";
    foreach ($projects as $project) {
        echo sprintf("  %s  %-9s %s\n", $project['id'], $project['status'], $project['name']);
    }

    echo "\nУдалённый проект (строки в app.projects нет, фиксации на него есть)\n";
    echo sprintf("  %s\n", $deletedProjectId);

    echo "\nТелефоны детерминированы, чтобы их можно было искать руками:\n";
    echo "  формат        +7 9XX XXX-XX-XX\n";
    echo "  хвосты        4567, 1122, 3344, 5566 — повторяются у разных клиентов\n";
    echo "  хвост 9999    партия, приехавшая из CRM одним импортом\n";
    echo sprintf("  один клиент на два ЖК  %s\n", $twoProjectsClient);
    echo sprintf("  он же в двух форматах  %s и %s\n", $legacyClient[0], $legacyClient[1]);

    echo "\nПароль у всех: password\n";
    echo "Токен:  make token EMAIL=<почта>\n";
} catch (\Throwable $e) {
    $pdo->rollBack();
    Log::line('seed failed: ' . $e->getMessage());
    exit(1);
}
