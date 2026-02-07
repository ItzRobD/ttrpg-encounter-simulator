package core

// hpStatusStub is a minimal implementation of HPStatus used for non-targetable entities like Lair.
type hpStatusStub struct{}

func NewHPStatusStub() HPStatus { return hpStatusStub{} }

func (hpStatusStub) GetHP() int           { return 0 }
func (hpStatusStub) GetMaxHP() int        { return 0 }
func (hpStatusStub) GetTempHP() int       { return 0 }
func (hpStatusStub) GetHPPct() int        { return 100 }
func (hpStatusStub) GetHPDifference() int { return 0 }
func (hpStatusStub) GetHitDie() DiceType  { return D0 }
