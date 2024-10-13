package models

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date,omitempty"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type TaskResponceSuccess struct {
	ID string `json:"id"`
}

type TaskResponceError struct {
	Error string `json:"error"`
}

type GetAllTaksksResponceSuccess struct {
	Tasks []*Task `json:"tasks"`
}
