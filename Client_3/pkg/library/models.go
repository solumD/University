package library

type Book struct {
	ID       int    `json:"id,omitempty"`
	Title    string `json:"title"`
	Year     int    `json:"year"`
	AuthorID int    `json:"author_id"`
}

type Author struct {
	ID      int    `json:"id,omitempty"`
	Name    string `json:"name"`
	Country string `json:"country"`
}

type Reader struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

type Loan struct {
	ID         int    `json:"id"`
	BookID     int    `json:"book_id"`
	UserID     int    `json:"user_id"`
	Status     string `json:"status"`
	BorrowDate string `json:"borrow_date"`
	ReturnDate string `json:"return_date"`
}

type LoanRequest struct {
	BookID int `json:"book_id"`
	UserID int `json:"user_id"`
}

type Page[T any] struct {
	Items  []T `json:"items"`
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}
