// /*********************************************************************************
// * Projeto:     Batedor
// * Componente:  Database - O cérebro e a memória do Batedor (Corrigido)
// *********************************************************************************/
package main

import (
	"database/sql"
	"fmt" // <-- ESTA LINHA FOI ADICIONADA
	"time"

	_ "github.com/mattn/go-sqlite3" // O driver do SQLite
)

const dbFile = "./batedor_history.db"

var db *sql.DB

// MetricRecord representa um único registro de dados históricos.
type MetricRecord struct {
	Timestamp time.Time
	Value     float64
}

// initDatabase abre a conexão com o banco de dados e cria a tabela se ela não existir.
func initDatabase() error {
	var err error
	db, err = sql.Open("sqlite3", dbFile)
	if err != nil {
		return err
	}

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS metrics (
		timestamp DATETIME NOT NULL,
		metric_name TEXT NOT NULL,
		value REAL NOT NULL,
		PRIMARY KEY (timestamp, metric_name)
	);
	`
	_, err = db.Exec(sqlStmt)
	return err
}

// logMetric salva uma nova métrica no banco de dados.
func logMetric(name string, value float64) {
	if db == nil {
		return
	}

	stmt, err := db.Prepare("INSERT INTO metrics(timestamp, metric_name, value) values(?,?,?)")
	if err != nil {
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(time.Now(), name, value)
}

// getMetricsForLast24h busca os dados históricos de uma métrica específica.
func getMetricsForLast24h(metricName string) ([]MetricRecord, error) {
	if db == nil {
		return nil, fmt.Errorf("banco de dados não inicializado")
	}

	rows, err := db.Query(`
		SELECT timestamp, value FROM metrics 
		WHERE metric_name = ? AND timestamp >= ? 
		ORDER BY timestamp ASC`,
		metricName, time.Now().Add(-24*time.Hour),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []MetricRecord
	for rows.Next() {
		var rec MetricRecord
		if err := rows.Scan(&rec.Timestamp, &rec.Value); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, nil
}