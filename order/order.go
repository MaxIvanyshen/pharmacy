package order

import (
    "pharmacy_system/entities"
    "time"
)

var count = 0

type Order struct {
    ID int `json:"id"`
    Meds []entities.Medicine `json:"meds"`
    Price int `json:"price"`
    Date string `json:"date"`
}

func NewOrder(meds []entities.Medicine) Order {
    count++
    price := 0
    for _, med := range meds {
        price += (med.Price * med.Count)
    }
    return Order{ID: count, Meds: meds, Price: price, Date: time.Now().Format("2006-01-02")}
}   

func (o Order) GetPrice() int {
    var price int
    for _, med := range o.Meds {
        price += (med.Price * med.Count)
    }
    return price
}

