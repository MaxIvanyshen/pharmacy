package dto

import (
    "pharmacy_system/entities"
    "pharmacy_system/order"
)

type OrderData struct {
    Orders []order.Order
}

type MedsData struct {
    Meds []entities.Medicine
}

type NewOrderData struct {
    Meds []entities.Medicine
    Order order.Order
}

