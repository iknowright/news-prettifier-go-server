package app

import (
	"database/sql"
	"fmt"
)

type news struct {
	ID    int     `json:"id"`
	Link  string  `json:"link"`
	ReturnLink string `json:"returnlink"`
}

func (n *news) getNews(db *sql.DB) error {
	return db.QueryRow("SELECT link, returnlink FROM news WHERE id=$1", n.ID).Scan(&n.Link, &n.ReturnLink)
}

func (n *news) updateNews(db *sql.DB) error {
	_, err := db.Exec("UPDATE news SET link=$1, returnlink=$2 WHERE id=$3", n.Link, n.ReturnLink, n.ID)
	return err
}

func (n *news) deleteNews(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM news WHERE id=$1", n.ID)
	return err
}

func (n *news) createNews(db *sql.DB) error {
	// postgres doesn't return the last inserted ID so this is the workaround
	err := db.QueryRow(
		"INSERT INTO news(link, returnlink) VALUES($1, $2) RETURNING id",
		n.Link, n.ReturnLink).Scan(&n.ID)
	return err
}

func getMultipleNews(db *sql.DB, start, count int) ([]news, error) {
	rows, err := db.Query("SELECT id, link, returnlink FROM news LIMIT $1 OFFSET $2", count, start)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	multipleNews := []news{}

	for rows.Next() {
		var n news
		if err := rows.Scan(&n.ID, &n.Link, &n.ReturnLink); err != nil {
			return nil, err
		}
		multipleNews = append(multipleNews, n)
	}

	return multipleNews, nil
}

type account struct {
	Username  string  `json:"username"`
	Password  string  `json:"password"`
	Email string `json:"email"`
	Size int `json:"size"`
	Color int `json:"color"`
 }

func (u *account) getAccount(db *sql.DB) error {
	return db.QueryRow("SELECT password, email, size, color FROM account WHERE username=$1", u.Username).Scan(&u.Password, &u.Email, &u.Size, &u.Color)
}

func (u *account) updateAccountSettings(db *sql.DB) error {
	_, err := db.Exec("UPDATE account SET size=$1, color=$2 WHERE username=$3", u.Size, u.Color, u.Username)
	return err
}

func (u *account) deleteAccount(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM account WHERE username=$1", u.Username)
	return err
}

func (u *account) createAccount(db *sql.DB) error {
	// postgres doesn't return the last inserted Username so this is the workaround
	err := db.QueryRow(
		"INSERT INTO account(username, password, email) VALUES($1, $2, $3) RETURNING username",
		u.Username, u.Password, u.Email).Scan(&u.Username)
	return err
}

type article struct {
	Article_ID  string  `json:"article_id"`
	Title  string  `json:"title"`
	Author string  `json:"author"`
	Content string  `json:"content"`
	Origin string  `json:"origin"`
	Username string  `json:"username"`
}

func (n *article) createArticle(db *sql.DB) error {
	// postgres doesn't return the last inserted ID so this is the workaround
	err := db.QueryRow(
		"INSERT INTO article(title, author, content, origin) VALUES($1, $2, $3, $4) RETURNING article_id",
		n.Title, n.Author, n.Content, n.Origin).Scan(&n.Article_ID)
	return err
}

func (n *article) getArticle(db *sql.DB) error {
	return db.QueryRow("SELECT title, author, content, origin FROM article WHERE article_id=$1", n.Article_ID).Scan(&n.Title, &n.Author, &n.Content, &n.Origin)
}

func (n *article) updateArticleUser(db *sql.DB , username string) error {
	_, err := db.Exec("UPDATE article SET username=$1 WHERE article_id=$2", username, n.Article_ID)
	return err
}

func getArticles(db *sql.DB, username string) ([]article, error) {
	fmt.Println(username)
	rows, err := db.Query("SELECT article_id, title FROM article WHERE username=$1", username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	articles := []article{}
	for rows.Next() {
		var p article
		if err := rows.Scan(&p.Article_ID, &p.Title); err != nil {
			return nil, err
		}
		// fmt.Printf("%+v\n", p);
		articles = append(articles, p)
	}

	return articles, nil
}

func deleteTimeoutArticle(db *sql.DB) error {
	fmt.Println("I am runnning delete timeout.")	
	_, err := db.Exec("DELETE FROM article WHERE username='' AND (CURRENT_TIMESTAMP - time) > INTERVAL '5 minutes'")

	return err
}