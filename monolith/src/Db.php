<?php
declare(strict_types=1);

namespace Broker\Monolith;

use PDO;

/**
 * Тонкая обёртка над PDO.
 *
 * Роль подключения — app_user: полные права на схему app, только select
 * на integration. Это не настройка «на всякий случай»: если код монолита
 * попробует записать в integration.fixations, база ответит permission
 * denied. Границу контуров держат гранты, а не то, что мы договорились
 * туда не писать.
 */
final class Db
{
    private static ?PDO $pdo = null;

    public static function pdo(): PDO
    {
        if (self::$pdo !== null) {
            return self::$pdo;
        }

        $dsn      = getenv('APP_DB_DSN') ?: 'pgsql:host=127.0.0.1;port=5432;dbname=broker';
        $user     = getenv('APP_DB_USER') ?: 'app_user';
        $password = getenv('APP_DB_PASSWORD') ?: 'app_user';

        $pdo = new PDO($dsn, $user, $password, [
            // Исключения вместо кодов возврата: молча проигнорированная
            // ошибка вставки — худшее, что может случиться с деньгами.
            PDO::ATTR_ERRMODE            => PDO::ERRMODE_EXCEPTION,
            PDO::ATTR_DEFAULT_FETCH_MODE => PDO::FETCH_ASSOC,
            // Настоящие prepared statements на стороне сервера,
            // а не подстановка строк драйвером.
            PDO::ATTR_EMULATE_PREPARES   => false,
        ]);

        // search_path задаём явно и здесь тоже: в запросах схема всё равно
        // пишется руками, но если кто-то забудет — уедет в app, а не в public.
        $pdo->exec("set search_path to app, public");

        self::$pdo = $pdo;

        return $pdo;
    }
}
