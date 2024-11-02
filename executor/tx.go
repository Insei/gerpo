package executor

import (
	"database/sql"
	"sync"
)

var txContextKey = &txContextKeyStruct{
	key: "tx",
}

type txContextKeyStruct struct {
	key string
}

type txData struct {
	mtx *sync.Mutex
	m   map[*sql.DB]*sql.Tx
}

func (t *txData) getTx(db *sql.DB) (*sql.Tx, bool) {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	tx, ok := t.m[db]
	return tx, ok
}

func (t *txData) setTx(db *sql.DB, tx *sql.Tx) {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	t.m[db] = tx
}

func newTxData() *txData {
	return &txData{
		mtx: &sync.Mutex{},
		m:   make(map[*sql.DB]*sql.Tx),
	}
}
