package main

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"pharmacy_system/db"
	"pharmacy_system/entities"
	"pharmacy_system/order"
	"pharmacy_system/dto"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

type Template struct {
    templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

var medsRepo db.MedsRepo
var ordersRepo db.OrderRepo

func getAllMedicine(c echo.Context) error {
    allMeds, err := medsRepo.GetAll()
    if err != nil {
        c.Error(err)
    }
    return c.Render(http.StatusOK, "table", dto.MedsData{Meds: allMeds})
}

func addMedicine(c echo.Context) error {
    var item entities.Medicine

    item.Name = c.FormValue("name")
    item.Price, _ = strconv.Atoi(c.FormValue("price"))
    item.ExpirationDate = c.FormValue("expirationDate")
    item.Count, _ = strconv.Atoi(c.FormValue("count"))
    item = entities.NewMedicine(item.Name, item.Price, item.ExpirationDate, item.Count)
    err := medsRepo.Save(item)  
    returnSlice, err := medsRepo.GetAll()
    if err != nil {
        c.Error(err)
    }
    return c.Render(http.StatusOK, "table", dto.MedsData{Meds: returnSlice})
}

func newOrder(c echo.Context) error {
    currentOrder.Date = time.Now().Format("2006-01-02")
    err := ordersRepo.Save(currentOrder)
    if err != nil {
        c.Error(err)
    }
    currentOrder = order.Order{}
    meds, err := medsRepo.GetAll()
    if err != nil {
        c.Error(err)
    }
    return c.Render(http.StatusOK, "new_order", dto.NewOrderData{Meds: meds, Order: currentOrder})
}

func getOrderById(c echo.Context) error {
    id, err := strconv.Atoi(c.FormValue("id"))
    if err != nil || id == 0 {
        return getAllOrders(c)
    }
    item, err := ordersRepo.FindById(id)
    if err != nil {
        c.Error(err)
    }
    var orders []order.Order
    orders = append(orders, item)
    return c.Render(http.StatusOK, "ordersTable", dto.OrderData{Orders: orders})
}

func deleteOrderById(c echo.Context) error {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.Error(errors.New("o holera, de ID?"))
    }
    err = ordersRepo.DeleteById(id)
    if err != nil {
        c.Error(err)
    }
    return getAllOrders(c)
}

func getOrdersByDate(c echo.Context) error {
    var foundOrders []order.Order
    date := c.FormValue("date")
    _, err := time.Parse("2006-01-02", date)
    if err != nil || date == "" {
        return getAllOrders(c)
    }

    foundOrders, _ = ordersRepo.FindByDate(date)
    return c.Render(http.StatusOK, "ordersTable", dto.OrderData{Orders: foundOrders})
}

func getAllOrders(c echo.Context) error {
    orders, err := ordersRepo.GetAll()
    if err != nil {
        c.Error(err)
    }
    return c.Render(http.StatusOK, "ordersTable", dto.OrderData{Orders: orders})
}

func getMedicineById(c echo.Context) error {
    id, err := strconv.Atoi(c.FormValue("id"))
    if err != nil || id == 0 {
        return getAllMedicine(c)
    }
    item, err := medsRepo.FindById(id)
    if err != nil {
        c.Error(err)
    }
    var meds []entities.Medicine
    meds = append(meds, item)
    return c.Render(http.StatusOK, "table", dto.MedsData{Meds: meds})
}

func getMedicineByName(c echo.Context) error {
    name := c.FormValue("name")
    if name == "" {
        return getAllMedicine(c)
    }
    items, err := medsRepo.FindByName(name)
    if err != nil {
        c.Error(err)
    }
    return c.Render(http.StatusOK, "table", dto.MedsData{Meds: items})
}

func getMedicineByNameInOrder(c echo.Context) error {
    name := c.FormValue("name")
    if name == "" {
        return getAllMedicine(c)
    }
    items, err := medsRepo.FindByName(name)
    if err != nil {
        c.Error(err)
    }
    return c.Render(http.StatusOK, "new_order", dto.NewOrderData{Meds: items, Order: currentOrder})
}

func addItemToOrder(c echo.Context) error {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.Error(err)
    }

    item, err := medsRepo.FindById(id)
    if err != nil {
        c.Error(err)
    }
    fmt.Println(item)

    item.Count = -1
    medsRepo.Save(item)

    exists := false
    for i, orderMeds := range currentOrder.Meds {
        if orderMeds.ID == item.ID {
           currentOrder.Meds[i].Count += 1 
           exists = true
        }
    }
    if !exists {
        item.Count = 1
        currentOrder.Meds = append(currentOrder.Meds, item)
    }

    meds, err := medsRepo.GetAll()
    if err != nil {
        c.Error(err)
    }
    currentOrder.Price += item.Price
    return c.Render(http.StatusOK, "new_order", dto.NewOrderData{Meds: meds, Order: currentOrder})
}
var currentOrder order.Order

func main() {
    t := &Template {
        templates: template.Must(template.ParseGlob("public/views/*.html")),
    }

    db.Connect()
    medsRepo = db.GetMedsRepo()
    ordersRepo = db.GetOrdersRepo()
    defer db.CloseConnection()
    
    e := echo.New()
    e.Renderer = t


    e.GET("/", func(c echo.Context) error {
        meds, err := medsRepo.GetAll()
        if err != nil {
            c.Error(err)
        } 
        return c.Render(http.StatusOK, "index", dto.MedsData{Meds: meds})
    })

    e.GET("/orders", func(c echo.Context) error {
        orders, err := ordersRepo.GetAll()
        if err != nil {
            panic(err)
        }
        return c.Render(200, "orders", dto.OrderData{Orders: orders})
    })

    e.GET("/orders/new", func(c echo.Context) error {
        meds, err := medsRepo.GetAll()
        if err != nil {
            panic(err)
        }
        currentOrder = order.Order{}
        return c.Render(200, "new_order", dto.NewOrderData{Meds: meds, Order: currentOrder})
    })

    e.DELETE("/api/meds", func(c echo.Context) error {
        err := medsRepo.RemoveExpired()
        allMeds, err := medsRepo.GetAll()
        if err != nil {
            c.Error(err)
        }
        return c.Render(http.StatusOK, "table", dto.MedsData{Meds: allMeds})
    })

    e.GET("/api/meds",getAllMedicine)
    e.POST("/api/meds", addMedicine)
    e.GET("/api/meds/id", getMedicineById)
    e.GET("/api/meds/name", getMedicineByName)
    e.GET("/api/orders/meds/name", getMedicineByNameInOrder)
    e.POST("/api/orders", newOrder)
    e.GET("/api/orders", getAllOrders)
    e.GET("/api/orders/id", getOrderById)
    e.DELETE("/api/orders/id/:id", deleteOrderById)
    e.GET("/api/orders/date", getOrdersByDate)
    e.POST("/api/orders/add_item/:id", addItemToOrder)

    e.Logger.Fatal(e.Start(":42069")) 
}
