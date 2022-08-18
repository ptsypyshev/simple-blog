package models

import (
	"fmt"
	"time"
)

type User struct {
	Id        int    `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	IsActive  bool   `json:"is_active"`
}

type Post struct {
	Id     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	UserId int    `json:"user_id"`
}

type Comment struct {
	Id     int       `json:"id"`
	Date   time.Time `json:"date"`
	Body   string    `json:"body"`
	UserId int       `json:"user_id"`
	PostId int       `json:"post_id"`
}

func (u User) String() string {
	return fmt.Sprintf("{\nID: %d\nUsername: %s\nPassword: %s\nFirstName: %s\nLastName: %s\nEmail: %s\nIsActive: %t\n}",
		u.Id, u.Username, u.Password, u.FirstName, u.LastName, u.Email, u.IsActive)
}

func (p Post) String() string {
	return fmt.Sprintf("{\nID: %d\nTitle: %s\nBody: %s\nUserId: %d\n}",
		p.Id, p.Title, p.Body, p.UserId)
}

func (c Comment) String() string {
	return fmt.Sprintf("{\nID: %d\nDate: %s\nBody: %s\nUserId: %d\nPostId: %d\n}",
		c.Id, c.Date, c.Body, c.UserId, c.PostId)
}
