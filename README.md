# transactor

Transactor is a simple pgx/v4 db transactor manager. Implementation is based on examples from [Clean Transactions in Hexagon](https://www.kaznacheev.me/posts/en/clean-transactions-in-hexagon/).

## Usage

### Installation
```
go get github.com/immannino/transactor
```

### Example

[See an example](./examples/main.go)

```golang
package main

import (
	"context"
	"log"

	"github.com/immannino/transactor"
	"github.com/jackc/pgx/v4/pgxpool"
)

type App struct {
	db   *pgxpool.Pool
	txer transactor.Transactor
}

func main() {
	db, _ := pgxpool.Connect(context.Background(), "postgres://admin:admin@localhost:5432")
	txer, _ := transactor.New(db)

	app := &App{db, txer}

	// do some stuff
	_ = app.txer.WithinTransaction(context.Background(), func(ctx context.Context) error {
		tx := transactor.ExtractTx(ctx)

		_, err := tx.Exec(ctx, "INSERT INTO users (name, phone) VALUES ('Spongebob', '80012345678')")
		if err != nil {
			log.Println("Error with txn")
			return err
		}

		log.Println("Successfully inserted users")
		return nil
	})
}
```
