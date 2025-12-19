package handshake

type provider interface {
	sendHello() error
	receiveHello() error
	sendMissingData() error
	receiveMissingData() error
}

func run[P provider](p P) error {
	if err := p.sendHello(); err != nil {
		return err
	}
	if err := p.receiveHello(); err != nil {
		return err
	}
	if err := p.sendMissingData(); err != nil {
		return err
	}
	if err := p.receiveMissingData(); err != nil {
		return err
	}
	return nil
}
