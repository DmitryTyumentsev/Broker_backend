<?php
declare(strict_types=1);

namespace Broker\Monolith;

use PhpAmqpLib\Connection\AMQPStreamConnection;
use PhpAmqpLib\Channel\AMQPChannel;
use PhpAmqpLib\Message\AMQPMessage;

/**
 * Шина между рантаймами.
 *
 * Договорённость по топологии:
 *   admin.events    — publisher монолит, consumer Go.
 *                     agency.approved, tariff.changed
 *   fixation.events — publisher Go, consumer монолит.
 *                     fixation.created, fixation.expired, ...
 *
 * Оба обменника topic: routing key имеет структуру, и подписчик берёт
 * либо конкретное событие, либо всю ветку через fixation.*. Direct не
 * подошёл бы — пришлось бы перечислять все ключи в каждом биндинге.
 *
 * Обменники durable: они переживают рестарт брокера. Недолговечный
 * обменник исчезает вместе с нодой, и публикация после рестарта уходит
 * в никуда без единой ошибки.
 */
final class Bus
{
    public const EXCHANGE_ADMIN    = 'admin.events';
    public const EXCHANGE_FIXATION = 'fixation.events';

    private static ?AMQPStreamConnection $connection = null;

    public static function channel(): AMQPChannel
    {
        if (self::$connection === null) {
            self::$connection = new AMQPStreamConnection(
                getenv('APP_AMQP_HOST') ?: '127.0.0.1',
                (int)(getenv('APP_AMQP_PORT') ?: 5672),
                getenv('APP_AMQP_USER') ?: 'broker',
                getenv('APP_AMQP_PASSWORD') ?: 'broker'
            );
        }

        $channel = self::$connection->channel();

        // Объявляем обменники при каждом подключении. Это идемпотентно:
        // если обменник уже есть с теми же параметрами, ничего не произойдёт.
        // Так ни publisher, ни consumer не зависят от того, кто поднялся первым.
        $channel->exchange_declare(self::EXCHANGE_ADMIN, 'topic', false, true, false);
        $channel->exchange_declare(self::EXCHANGE_FIXATION, 'topic', false, true, false);

        return $channel;
    }

    /**
     * @param array<string,mixed> $payload
     */
    public static function publish(string $exchange, string $routingKey, array $payload): void
    {
        $channel = self::channel();

        $message = new AMQPMessage(
            json_encode($payload, JSON_UNESCAPED_UNICODE | JSON_THROW_ON_ERROR),
            [
                'content_type'  => 'application/json',
                // delivery_mode=2 — сообщение пишется на диск. Без этого
                // всё, что лежало в очереди, пропадёт при рестарте брокера.
                'delivery_mode' => AMQPMessage::DELIVERY_MODE_PERSISTENT,
                // message_id нужен потребителю, чтобы отбрасывать дубли:
                // at-least-once доставка означает, что одно и то же
                // сообщение придёт дважды, и это норма, а не сбой.
                'message_id'    => bin2hex(random_bytes(16)),
                'timestamp'     => time(),
            ]
        );

        $channel->basic_publish($message, $exchange, $routingKey);
        $channel->close();

        Log::line("published {$exchange} / {$routingKey}");
    }
}
