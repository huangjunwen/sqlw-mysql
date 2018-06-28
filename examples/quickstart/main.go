package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/huangjunwen/sqlw-mysql/examples/quickstart/models"
)

func main() {
	ctx := context.Background()

	// Open db.
	db, err := sql.Open("mysql", "root:123456@tcp(localhost:16033)/dev?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Open a single connection.
	conn, err := db.Conn(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	{
		log.Printf("\n")
		log.Printf(">>>> Iter all user and its associated employee\n")

		slice, err := models.AllUserEmployees(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}

		for _, result := range slice {
			user := result.User
			empl := result.Empl

			if empl.Valid() {
				log.Printf("User %+q (age %d) is an employee, sn: %+q\n", user.Name, result.Age.Uint64, empl.EmployeeSn)
			} else {
				log.Printf("User %+q (age %d) is not an employee\n", user.Name, result.Age.Uint64)
			}
		}
	}

	{
		log.Printf("\n")
		log.Printf(">>>> Iter subordinate\n")

		slice, err := models.SubordinatesBySuperiors(ctx, conn, 1, 2, 3, 4, 5, 6, 7)
		if err != nil {
			log.Fatal(err)
		}

		superiors, groups := slice.GroupBySuperior()
		for i, superior := range superiors {
			subordinates := groups[i].DistinctSubordinate()

			if len(subordinates) == 0 {
				log.Printf("Employee %+q has no subordinate.\n", superior.EmployeeSn)
			} else {
				log.Printf("Employee %+q has the following subordinates:\n", superior.EmployeeSn)
				for _, subordinate := range subordinates {
					log.Printf("\t%+q\n", subordinate.EmployeeSn)
				}
			}

		}

	}

	{
		log.Printf("\n")
		log.Printf(">>>> Query user by different condition\n")

		{
			slice, err := models.UsersByCond(ctx, conn, 0, "Zombie", time.Time{}, 1)
			if err != nil {
				log.Fatal(err)
			}
			for _, result := range slice {
				log.Printf("id: %d, name: %+q, femal: %v, birthday: %v", result.Id, result.Name, result.Female, result.Birthday)
			}
		}

		{
			slice, err := models.UsersByCond(ctx, conn, 0, "", time.Date(1992, time.Month(2), 2, 0, 0, 0, 0, time.UTC), 10)
			if err != nil {
				log.Fatal(err)
			}
			for _, result := range slice {
				log.Printf("id: %d, name: %+q, femal: %v, birthday: %v", result.Id, result.Name, result.Female, result.Birthday)
			}
		}

		{
			slice, err := models.UsersByCond(ctx, conn, 1, "", time.Date(1992, time.Month(2), 2, 0, 0, 0, 0, time.UTC), 10)
			if err != nil {
				log.Fatal(err)
			}
			for _, result := range slice {
				log.Printf("id: %d, name: %+q, femal: %v, birthday: %v", result.Id, result.Name, result.Female, result.Birthday)
			}
		}

	}

}
