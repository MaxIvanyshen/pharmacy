package db

import (
	"database/sql"
	"errors"
	"log"
	"pharmacy_system/entities"
	"pharmacy_system/order"
	"strings"
    "time"

	_ "github.com/lib/pq"
)

type MedsRepo struct {
    db *sql.DB
}

func (repo *MedsRepo) RemoveExpired() error {
    meds, err := repo.GetAll()
    if err != nil {
        return err
    }
    for _, m := range meds {
        expDate, err := time.Parse("2006-01-02", m.ExpirationDate)
        if err != nil {
            return err
        }
        if expDate.Before(time.Now()) {
            _, err := db.Query("DELETE FROM meds WHERE id = $1", m.ID)
            if err != nil {
                return err
            }
            continue
        }
    }
    return nil
}

func (repo *MedsRepo) GetAll() ([]entities.Medicine, error) {
    var meds []entities.Medicine

    rows, err := db.Query("SELECT * FROM meds")  
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    for rows.Next() {
        var m entities.Medicine
        if err := rows.Scan(&m.ID, &m.Name, &m.Price, &m.ExpirationDate, &m.Count); err != nil {
            return meds, err
        }
        meds = append(meds, m)
    }

    return meds, nil
}

func (repo *MedsRepo) FindById(id int) (entities.Medicine, error) {
    var med entities.Medicine
    med.ID = id
    rows, err := db.Query("SELECT name, price, expirationdate FROM meds WHERE id = $1", id)
    if err != nil {
        return med, err
    }
    defer rows.Close()
    
    if rows.Next() {
        if err = rows.Scan(&med.Name, &med.Price, &med.ExpirationDate); err != nil {
            return med, err
        }
        return med, nil
    }
    
    return entities.Medicine{}, errors.New("no medicine")
}

func (repo *MedsRepo) FindByName(name string) ([]entities.Medicine, error) {
    var meds []entities.Medicine

    allMeds, err := repo.GetAll()
    if err != nil {
        return make([]entities.Medicine, 0), nil
    }

    for i, item := range allMeds {
        if strings.Contains(item.Name, name) {
            meds = append(meds, allMeds[i])
        }
    }

    return meds, nil
}

func (repo *MedsRepo) Save(med entities.Medicine) error {
    var count int
    rows, err := repo.db.Query("SELECT id FROM meds ORDER BY id DESC LIMIT 1;")
    if err != nil {
        return err
    }
    if rows.Next() {
        err = rows.Scan(&count)
        if err != nil {
            return err
        }
        med.ID = count + 1; 
    } else {
        med.ID = 1
    }
    meds, err := repo.GetAll()
    if err != nil {
        return err
    }
    for _, item := range meds {
        if item.Name == med.Name && med.ExpirationDate == item.ExpirationDate && item.Price == med.Price {
            _, err = repo.db.Query("UPDATE meds SET count = $1 WHERE id = $2;", item.Count + med.Count, item.ID)                          
            return err
        }
    }

    _, err = repo.db.Query(
        "INSERT INTO meds(id, name, price, expirationdate, count) VALUES($1, $2, $3, $4, $5)", 
        med.ID, 
        med.Name, 
        med.Price, 
        med.ExpirationDate, 
        med.Count,
    )
    return err
}

type OrderRepo struct {
    db *sql.DB
}

func (repo *OrderRepo) GetAll() ([]order.Order, error) {
    var orders []order.Order

    rows, err := db.Query("SELECT * FROM orders")  
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    for rows.Next() {
        var o order.Order
        if err := rows.Scan(&o.ID, &o.Price, &o.Date); err != nil {
            return orders, err
        }
        o, err = repo.parseOrderData(o)
        if err != nil {
            return orders, err
        }
        orders = append(orders, o)
    }

    return orders, nil
}

func (repo *OrderRepo) DeleteById(id int) error {
    _, err := repo.db.Query("DELETE FROM orders WHERE id = $1", id)
    if err != nil {
        return err
    }

    _, err = repo.db.Query("DELETE FROM meds_in_order WHERE order_id = $1", id)
    if err != nil {
        return err
    }
    return nil
}

func (repo *OrderRepo) parseOrderData(order order.Order) (order.Order, error) {
        var meds []entities.Medicine
        relationRows, err := db.Query("SELECT meds_id, meds_count FROM meds_in_order WHERE order_id = $1", order.ID)
        if err != nil {
            return order, err
        }
        defer relationRows.Close()
        
        var count int
        var meds_id int

        for relationRows.Next() {
            if err = relationRows.Scan(&meds_id, &count); err != nil {
                return order, err
            }
            
            var name string
            var price int
            var expDate string

            medsDataRows, err := db.Query("SELECT name, price, expirationdate FROM meds WHERE id = $1;", meds_id)
            if err != nil {
                return order, err
            }
            defer medsDataRows.Close()

            if medsDataRows.Next() {
               if err = medsDataRows.Scan(&name, &price, &expDate); err != nil {
                   return order, err
               }
               meds = append(meds, entities.Medicine{
                   ID: meds_id, 
                   Name: name, 
                   Price: price, 
                   ExpirationDate: expDate, 
                   Count: count,
               })
            }
        }          
        order.Meds = meds    

        return order, nil
}

func (repo *OrderRepo) Save(order order.Order) error {
    var count int
    rows, err := repo.db.Query("SELECT id FROM orders ORDER BY id DESC LIMIT 1;")
    if err != nil {
        return err
    }
    if rows.Next() {
        err = rows.Scan(&count)
        if err != nil {
            return err
        }
        order.ID = count + 1; 
    } else {
        order.ID = 1
    }
    medsRepo := GetMedsRepo()
    allMeds, err := medsRepo.GetAll()
    if err != nil {
        return err
    }

    for _, orderItem := range order.Meds {
        for _, meds := range allMeds {
            if meds.Name == orderItem.Name && meds.ExpirationDate == orderItem.ExpirationDate && meds.Price == orderItem.Price {
                if meds.Count >= orderItem.Count {
                    count := orderItem.Count
                    _, err = db.Query("INSERT INTO meds_in_order(order_id, meds_id, meds_count) VALUES($1, $2, $3)", order.ID, meds.ID, count)
                    if err != nil {
                            return errors.New("couldnt save orders meds")
                    }
                } else {
                    return errors.New("Not enough values in stock")
                }
            }
        }
    }

    _, err = db.Query("INSERT INTO orders(id, price, date) VALUES($1, $2, $3)", order.ID, order.Price, order.Date)
    if err != nil {
        return errors.New("couldnt save order")
    }

    return nil
}

func (repo *OrderRepo) FindById(id int) (order.Order, error) {
    var o order.Order
    o.ID = id
    rows, err := db.Query("SELECT price, date FROM orders WHERE id = $1", id)
    if err != nil {
        return o, err
    }
    defer rows.Close()
    
    if rows.Next() {
        if err = rows.Scan(&o.Price, &o.Date); err != nil {
            return o, err
        }
        o, err = repo.parseOrderData(o)  
        if err != nil {
            return o, nil
        }
    }

    return o, nil
}

func (repo *OrderRepo) FindByDate(date string) ([]order.Order, error) {
    var orders []order.Order
    rows, err := db.Query("SELECT id, price FROM orders WHERE date = $1", date)
    if err != nil {
        return orders, err
    }
    defer rows.Close()
    
    for rows.Next() {
        var o order.Order
        o.Date = date
        if err = rows.Scan(&o.ID, &o.Price); err != nil {
            return orders, err
        }
        o, err = repo.parseOrderData(o)  
        if err != nil {
            return orders, nil
        }
        orders = append(orders, o)
    }

    return orders, nil
}

var db *sql.DB
var err error

func Connect() { 
    connStr := "postgres://postgres:@localhost/meds?sslmode=disable"
    db, err = sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal(err)
    }
}

func CloseConnection() {
    db.Close()
}

func GetMedsRepo() MedsRepo {
    return MedsRepo {db: db}
}

func GetOrdersRepo() OrderRepo {
    return OrderRepo {db: db}
}
