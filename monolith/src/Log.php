<?php
declare(strict_types=1);

namespace Broker\Monolith;

/**
 * Лог в stderr одной строкой.
 *
 * stderr, а не файл: в контейнере логи собирает docker, а файл внутри
 * контейнера никто никогда не прочитает.
 */
final class Log
{
    public static function line(string $message): void
    {
        $timestamp = (new \DateTimeImmutable('now'))->format('Y-m-d H:i:s');
        fwrite(STDERR, "[monolith] {$timestamp} {$message}\n");
    }
}
