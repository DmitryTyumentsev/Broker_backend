<?php
declare(strict_types=1);

/**
 * Консьюмер fixation.* — сторона монолита.
 *
 * Go публикует события фиксации, монолит на них реагирует: письмо брокеру,
 * запись в CRM-задачи, пересчёт отчётов. Здесь всё это сведено к строчке
 * в логе — задача заглушки в том, чтобы связь между контурами была видна
 * и проверяема, а не в том, чтобы слать почту.
 *
 * В настоящем Laravel это был бы отдельный воркер очереди
 * (php artisan queue:work) с job-классом на каждое событие.
 */

require __DIR__ . '/../vendor/autoload.php';

use Broker\Monolith\Bus;
use Broker\Monolith\Log;
use PhpAmqpLib\Message\AMQPMessage;

$channel = Bus::channel();

// Очередь именованная и durable, а не эксклюзивная временная.
// Эксклюзивная умирает вместе с процессом, и всё, что Go опубликовал,
// пока монолит перезапускался, теряется без следа.
$queue = 'monolith.fixation.events';
$channel->queue_declare($queue, false, true, false, false);

// Одна подписка на всю ветку: fixation.created, fixation.expired,
// fixation.transferred и всё, что появится дальше. Новое событие
// не потребует правки биндинга.
$channel->queue_bind($queue, Bus::EXCHANGE_FIXATION, 'fixation.*');

// prefetch=1: брокер не отдаёт следующее сообщение, пока не подтверждено
// текущее. Без этого он вывалит всю очередь в память одного консьюмера,
// и второй экземпляр будет простаивать.
$channel->basic_qos(0, 1, false);

Log::line("consumer started, waiting for fixation.* on {$queue}");

$channel->basic_consume(
    $queue,
    '',
    false,
    false, // no_ack=false: подтверждаем руками, после обработки
    false,
    false,
    static function (AMQPMessage $message): void {
        $routingKey = $message->getRoutingKey();
        $payload    = json_decode($message->getBody(), true);

        if (!is_array($payload)) {
            // Битое тело повторять бессмысленно: оно не станет валиднее.
            // reject без requeue, иначе сообщение будет крутиться вечно
            // и забьёт консьюмер (poison message).
            Log::line("rejected malformed message on {$routingKey}");
            $message->reject(false);
            return;
        }

        Log::line(sprintf(
            'отправил письмо брокеру: событие=%s фиксация=%s агентство=%s',
            $routingKey,
            (string)($payload['fixation_id'] ?? '-'),
            (string)($payload['agency_id'] ?? '-')
        ));

        // ack только после успешной обработки. Подтвердить заранее —
        // значит потерять событие, если обработка упадёт посередине.
        $message->ack();
    }
);

while ($channel->is_consuming()) {
    $channel->wait();
}
