package requests

import "seminarska/internal/data/storage"

type task struct {
	job func(database *storage.AppDatabase)
}
