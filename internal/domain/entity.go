package domain

type URL struct {
    ID        string
    Original   string
    ShortCode  string
    CreatedAt  int64
    UpdatedAt  int64
}

type Click struct {
    ID        string
    ShortCode string
    IPAddress string
    UserAgent string
    ClickedAt int64
}
