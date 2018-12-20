package repository

type WlsDatabase interface {
	Migrate() error
	FlavorRepository() FlavorRepository
	ImageRepository() ImageRepository
}
