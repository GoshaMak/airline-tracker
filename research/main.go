package main

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	start = 10_000
	step  = 5_000
	stop  = 100_000
	reps  = 100
)

func main() {
	ctx := context.Background()
	files := [...]string{
		"res/flights_h.txt",
		"res/flights_b.txt",
		"res/flights.txt",
	}
	queries := [...]string{
		`explain analyze
		select * from flights`,
		`explain analyze
		select * from flights`,
		`explain analyze
		select * from flights`,
	}
	pre := [...]string{
		"create index flights_id_hash_idx on flights using hash (id)",
		"create index flights_id_btree_idx on flights using btree (id)",
		"",
	}
	after := [...]string{
		"drop index flights_id_hash_idx",
		"drop index flights_id_btree_idx",
		"",
	}

	pool, err := NewPostgresPool()
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := CloseConnection(pool); err != nil {
			slog.Error("can't close connection", "err", err)
		}
	}()

	for i := range pre {
		if len(pre[i]) > 0 {
			if _, err := pool.Exec(ctx, pre[i]); err != nil {
				slog.Error("create index", "query", pre[i], "err", err)
				return
			}
		}
		for s := start; s <= stop; s += step {
			slog.Info("starting measurement", "variant", i, "flights", s)
			Generate(pool, s)
			Measure(pool, files[i], queries[i])
		}
		if len(after[i]) > 0 {
			if _, err := pool.Exec(ctx, after[i]); err != nil {
				slog.Error("drop index", "query", after[i], "err", err)
				return
			}
		}
	}
}

func Measure(conn *pgxpool.Pool, fileName, query string) {
	ctx := context.Background()

	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		slog.Error("open file", "err", err)
		return
	}
	defer file.Close()

	wr := bufio.NewWriter(file)
	defer func() {
		if err := wr.Flush(); err != nil {
			slog.Error("flush", "err", err)
		}
	}()

	hotup(ctx, conn)

	flightsSize := sizeFlights(ctx, conn)

	for range reps {
		rows, err := conn.Query(ctx, query)
		if err != nil {
			slog.Error("explain analyze", "err", err)
			return
		}

		var planning float64
		var execution float64
		var hasPlanning bool
		var hasExecution bool

		for rows.Next() {
			var line string

			if err := rows.Scan(&line); err != nil {
				rows.Close()
				slog.Error("scan", "err", err)
				return
			}

			slog.Info("line: " + line)

			if value, err := ExtractPlanningTime(line); err == nil {
				planning = value
				hasPlanning = true
			}

			if value, err := ExtractExecutionTime(line); err == nil {
				execution = value
				hasExecution = true
			}
		}

		if err := rows.Err(); err != nil {
			rows.Close()
			slog.Error("rows", "err", err)
			return
		}

		rows.Close()

		if hasPlanning && hasExecution {
			if _, err := fmt.Fprintf(wr, "%d:%f:%f\n", flightsSize, planning, execution); err != nil {
				slog.Error("write", "err", err)
				return
			}
		} else {
			slog.Error(
				"planning or execution time not found",
				"hasPlanning", hasPlanning,
				"hasExecution", hasExecution,
			)
		}
	}

	slog.Info("FINISHED")
}

func hotup(ctx context.Context, conn *pgxpool.Pool) {
	_, err := conn.Exec(ctx, `analyze flights`)
	if err != nil {
		slog.Error("hotup flights", "err", err)
		panic(err)
	}
}

func sizeFlights(ctx context.Context, conn *pgxpool.Pool) int {
	var size int

	q := `
		select count(*)
		from flights
	`

	err := conn.QueryRow(ctx, q).Scan(&size)
	if err != nil {
		slog.Error("query row", "err", err)
		panic(err)
	}

	return size
}
