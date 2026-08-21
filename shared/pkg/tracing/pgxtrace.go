package tracing

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// QueryTracer превращает каждый запрос к PostgreSQL в спан.
//
// Зачем. Спаны у нас есть только на границах: HTTP-запрос в partnerapi и
// RPC в fixationservice. Внутри — чёрный ящик. Когда наружу прилетает 500,
// по логам видно «какой сервис», но не видно «на каком запросе к базе».
//
// С этим трейсером в Jaeger видна вся лестница:
//
//	POST /api/v1/fixations                    partnerapi   42ms
//	  fixation.v1.FixationService/NewFixation fixationsvc  38ms
//	    pg.query  select ... from app.projects              2ms
//	    pg.query  select ... from app.users                 1ms
//	    pg.query  insert into integration.fixations         3ms  ← error
//
// Отдельно полезно при «context deadline exceeded»: сразу видно, какой
// запрос съел время, вместо гадания.
//
// Аргументы запроса в спан НЕ попадают намеренно. В них ездят телефон
// и хэш телефона — персональные данные, и место им не в трейсах, которые
// хранятся отдельно и живут дольше логов.
type QueryTracer struct {
	tracer trace.Tracer
}

func NewQueryTracer(serviceName string) *QueryTracer {
	serviceName = strings.TrimSpace(serviceName)
	if serviceName == "" {
		serviceName = "postgres"
	}

	return &QueryTracer{tracer: otel.Tracer(serviceName)}
}

// TraceQueryStart открывает спан. pgx сам передаст возвращённый контекст
// в TraceQueryEnd, поэтому хранить спан где-то ещё не нужно.
func (t *QueryTracer) TraceQueryStart(
	ctx context.Context,
	_ *pgx.Conn,
	data pgx.TraceQueryStartData,
) context.Context {
	ctx, _ = t.tracer.Start(
		ctx,
		"pg.query "+firstLine(data.SQL),
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.statement", squashWhitespace(data.SQL)),
		),
	)

	return ctx
}

func (t *QueryTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	if data.Err != nil {
		span.RecordError(data.Err)
		span.SetStatus(otelcodes.Error, data.Err.Error())

		return
	}

	// Сколько строк затронул запрос. Ноль на insert — это не ошибка базы,
	// а бизнес-факт: место занято, ON CONFLICT сработал. Отличить одно
	// от другого по логам иначе нечем.
	span.SetAttributes(
		attribute.String("db.operation", data.CommandTag.String()),
		attribute.Int64("db.rows_affected", data.CommandTag.RowsAffected()),
	)
}

// firstLine достаёт из запроса первую содержательную строку — она попадёт
// в имя спана. Целый SQL в имени сделает список спанов нечитаемым.
func firstLine(sql string) string {
	const limit = 60

	sql = squashWhitespace(sql)
	if len(sql) > limit {
		return sql[:limit] + "…"
	}

	return sql
}

func squashWhitespace(sql string) string {
	return strings.Join(strings.Fields(sql), " ")
}
