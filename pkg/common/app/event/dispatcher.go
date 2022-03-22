package event

type Handler interface {
	Handle(event Event) error
}

type HandlerFunc func(event Event) error

func (f HandlerFunc) Handle(event Event) error {
	return f(event)
}

type Dispatcher interface {
	Dispatch(event Event) error
	Subscribe(handler Handler)
}

type dispatcher struct {
	handlers []Handler
}

func (d *dispatcher) Dispatch(event Event) error {
	for _, handler := range d.handlers {
		err := handler.Handle(event)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *dispatcher) Subscribe(handler Handler) {
	d.handlers = append(d.handlers, handler)
}

func NewEventDispatcher() Dispatcher {
	return &dispatcher{}
}
