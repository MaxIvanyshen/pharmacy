package entities

var medsCount int = 0

type Medicine struct {
    ID int `json:"id"`
    Name string `json:"name"`
    Price int `json:"price"`
    ExpirationDate string `json:"expirationDate"`
    Count int `json:"count"`
}

func NewMedicine(name string, price int, expirationDate string, count int) Medicine {
    medsCount += 1
    return Medicine{
        ID: medsCount,
        Name: name,
        Price: price,
        ExpirationDate: expirationDate,
        Count: count,
    }
}

