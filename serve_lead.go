package twig

import "context"

type Lead struct {
	t     *Twig
	works []Server
}

func (l *Lead) Attach(t *Twig) {
	l.t = t
}

func (l *Lead) Start() error {
	for _, s := range l.works {
		if err := s.Start(); err != nil {
			l.t.Logger.Println(err)
		}
	}
	return nil
}

func (l *Lead) Shutdown(ctx context.Context) error {
	for _, s := range l.works {
		if err := s.Shutdown(ctx); err != nil {
			l.t.Logger.Println(err)
		}
	}

	return nil
}

func (l *Lead) AddServer(servers ...Server) {
	l.works = append(l.works, servers...)
	for _, s := range l.works {
		s.Attach(l.t)
	}
}
