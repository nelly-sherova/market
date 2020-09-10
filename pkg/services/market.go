package services

import (
	"context"
	"errors"
	errors2 "github.com/nelly-sherova/market/pkg/errors"
	"github.com/nelly-sherova/market/pkg/models"
	"github.com/jackc/pgx/v4/pgxpool"
)

type NellyMarket struct {
	pool *pgxpool.Pool
}

func NewNellyMarket(pool *pgxpool.Pool) *NellyMarket {
	if pool == nil {
		panic(errors.New("pool can't be nil"))
	}
	return &NellyMarket{pool: pool}
}

//---------------------DB Tables ----------------------
func (receiver *NellyMarket) Start() {
	conn, err := receiver.pool.Acquire(context.Background())
	if err != nil {
		panic(errors.New("can't create database"))
	}
	defer conn.Release()

	_, err = conn.Exec(context.Background(), `
	Create table if not exists prices (
	id BIGSERIAL PRIMARY KEY,
	name TEXT NOT NULL,
	category TEXT NOT NULL,
	price INT,
	removed boolean DEFAULT FALSE
 );`)
	if err != nil {
		panic(errors.New("CAN'T Creat DB table for PRICES"))
	}

	_, err = conn.Exec(context.Background(), `
	Create table if not exists clients (
	id BIGSERIAL PRIMARY KEY,
	client TEXT NOT NULL,
	city TEXT NOT NULL,
	region TEXT NOT NULL
 );`)
	if err != nil {
		panic(errors.New("CAN'T Creat DB table for CLIENTS"))
	}

	_, err = conn.Exec(context.Background(), `
	Create table if not exists sales (
	id BIGSERIAL PRIMARY KEY,
	date date default now(),
	product TEXT NOT NULL,
	count int,
	sum int,
	client TEXT NOT NULL
 );`)
	if err != nil {
		panic(errors.New("CAN'T Creat DB table for SALES"))
	}

}

//---------------------Products List ----------------------
func (receiver *NellyMarket) ProductsList() (list []models.Prices, err error) {
	conn, err := receiver.pool.Acquire(context.Background())
	if err != nil {
		return nil, errors2.QueryErrors("can't execute pool: ", err)
	}
	defer conn.Release()

	rows, err := conn.Query(context.Background(), `SELECT id, name, category, price FROM prices where removed = false`)
	if err != nil {
		return nil, errors2.QueryErrors("can't select prices ", err)
	}

	defer rows.Close()
	list = make([]models.Prices, 0)
	for rows.Next() {
		item := models.Prices{}

		err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.Price)
		if err != nil {
			return nil, errors2.QueryErrors("can't scan ", err)
		}
		list = append(list, item)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (receiver *NellyMarket) AddProducts(prices models.Prices) (err error) {
	acquire, err := receiver.pool.Acquire(context.Background())
	if err != nil {
		return errors2.QueryErrors("Can't save ", err)
	}
	defer acquire.Release()
	_, err = acquire.Exec(context.Background(), `INSERT INTO prices(name, category, price) values ($1, $2, $3)`, prices.Name, prices.Category, prices.Price)
	if err != nil {
		return errors2.QueryErrors("CAN'T save product ", err)
	}
	return nil
}

func (receiver *NellyMarket) RemoveById(id int) (err error) {
	remove, err := receiver.pool.Acquire(context.Background())
	if err != nil {
		return errors2.QueryErrors("can't execute pool: ",err)
	}
	defer remove.Release()
	_, err = remove.Exec(context.Background(), "UPDATE prices SET removed = true WHERE id = $1", id)
	if err != nil {
		return errors2.QueryErrors("can't remove : ",err)
	}
	return nil
}

func (receiver *NellyMarket) AddSalesInDB(sales models.Sales) (err error) {
	add, err := receiver.pool.Acquire(context.Background())
	if err != nil {
		return errors2.QueryErrors("Can't add clients to DB ", err)
	}
	defer add.Release()
	_, err = add.Exec(context.Background(), `INSERT INTO sales(client, product, count) VALUES ($1,$2, $3)`, sales.Client, sales.Product, sales.Count)
	if err != nil {
		return errors2.QueryErrors("CAN'T save sales ", err)
	}
	return nil
}

func (receiver *NellyMarket) SalesList() (list []models.Sales, err error) {
	conn, err := receiver.pool.Acquire(context.Background())
	if err != nil {
		return nil, errors2.QueryErrors("CAN'T conn ",err)
	}
	defer conn.Release()

	rows, err := conn.Query(context.Background(), `SELECT s.id, s.client, s.count,coalesce( p.category, 'котегория не выбрана'),s.product, coalesce(s.count*p.price , 0 )as sum from sales s left join  prices p on s.product = p.name;`)
	if err != nil {
		return nil, errors2.QueryErrors("Can't select sales ",err)
	}
	defer rows.Close()

	list = make([]models.Sales, 0)
	for rows.Next() {
		item := models.Sales{}
		err := rows.Scan(&item.ID,&item.Client, &item.Count, &item.Category,&item.Product, &item.Sum)
		if err != nil {
			return nil,errors2.QueryErrors("Can't scan ",err)
		}
		list = append(list, item)
	}
	err = rows.Err()
	if err != nil {
		return nil,err
	}
	return list, nil
}

