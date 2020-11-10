package cache

// ListerInterface is tracked to handle late initialization of the controller
type ListerInterface interface {
	List(options interface{}) ([]interface{}, error)
}

// ListFunc knows how to list resources
type ListFunc func(options interface{}) ([]interface{}, error)

// Lister is defined to invoke ListFunc
type Lister struct {
	ListFunc ListFunc
}

// List a set of apiServer resources
func (lw *Lister) List(options interface{}) ([]interface{}, error) {
	return lw.ListFunc(options)
}
