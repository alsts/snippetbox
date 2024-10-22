package models

import (
	"database/sql"
	"errors"
	"time"
)

// Define a Snippet type to hold the data for an individual snippet. Notice how
// the fields of the struct correspond to the fields in our MySQL snippets
// table?
type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

// Define a SnippetModel type which wraps a sql.DB connection pool.
type SnippetModel struct {
	DB *sql.DB
}

func (m *SnippetModel) Insert(title string, content string, expires int) (int, error) {
	stmt := `INSERT INTO snippets (title, content, created, expires) 
	VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`

	result, err := m.DB.Exec(stmt, title, content, expires)
	if err != nil {
		return 0, err
	}
	// Use the LastInsertId() method on the result to get the ID of our
	// newly inserted record in the snippets table.
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	// The ID returned has the type int64, so we convert it to an int type
	// before returning.
	return int(id), nil
}

// This will return a specific snippet based on its id.
func (m *SnippetModel) Get(id int) (*Snippet, error) {
	s := &Snippet{}

	err := m.DB.QueryRow(
		`SELECT id, title, content, created, expires 
		 FROM snippets WHERE expires > UTC_TIMESTAMP() AND id = ?`, id,
	).Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	} // If everything went OK then return the Snippet object.
	return s, nil
}

// This will return the 10 most recently created snippets.
func (m *SnippetModel) Latest() ([]*Snippet, error) {
	stmt := `SELECT id, title, content, created, expires FROM snippets 
    WHERE expires > UTC_TIMESTAMP() ORDER BY id DESC LIMIT 10`

	rows, err := m.DB.Query(stmt) 
    if err != nil { 
        return nil, err 
    } 

	// We defer rows.Close() to ensure the sql.Rows resultset is 
    // always properly closed before the Latest() method returns. This defer 
    // statement should come *after* you check for an error from the Query() 
    // method. Otherwise, if Query() returns an error, you'll get a panic 
    // trying to close a nil resultset. 
	defer rows.Close() 

	snippets := []*Snippet{} 

	for rows.Next() { 
        // Create a pointer to a new zeroed Snippet struct. 
        s := &Snippet{} 
        // Use rows.Scan() to copy the values from each field in the row to the 
        // new Snippet object that we created. Again, the arguments to row.Scan() 
        // must be pointers to the place you want to copy the data into, and the 
        // number of arguments must be exactly the same as the number of 
        // columns returned by your statement. 
        err = rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires) 
        if err != nil { 
            return nil, err 
        } 
        // Append it to the slice of snippets. 
        snippets = append(snippets, s) 
    } 

	if err = rows.Err(); err != nil { 
        return nil, err 
    } 

	return snippets, nil
}
