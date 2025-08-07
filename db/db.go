package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func Connect(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`
      create table if not exists todos (
        id integer primary key autoincrement,
        text text not null,
        priority tinyint not null default 2,  -- 1 = low, 2 = medium, 3 = high 
        status tinyint not null default 1,    -- 1 = pending, 2 = processing, 3 = done
        due datetime,
        tag text,
        created_at datetime default current_timestamp,
        updated_at datetime default current_timestamp
      );
    `)
	return db, err
}
