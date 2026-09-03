package model

import (
	"fmt"

	"github.com/tweecore/dsfinvk"
	"github.com/tweecore/dsfinvk/schema"
)

// amounts is a Brutto, Netto and USt triple.
type amounts struct {
	brutto dsfinvk.Decimal
	netto  dsfinvk.Decimal
	ust    dsfinvk.Decimal
}

// add sums o into a.
func (a *amounts) add(o amounts) error {
	var err error
	if a.brutto, err = a.brutto.Add(o.brutto); err != nil {
		return err
	}
	if a.netto, err = a.netto.Add(o.netto); err != nil {
		return err
	}
	a.ust, err = a.ust.Add(o.ust)
	return err
}

// gvKey groups businesscases.csv rows. Spec 2.4 p.80.
type gvKey struct {
	typ        schema.GVTyp
	name       string
	agentur    int64
	schluessel int64
}

// payKey groups payment.csv rows. Spec 2.4 p.82.
type payKey struct {
	typ  schema.ZahlartTyp
	name string
}

// totals are the payment sums of cashpointclosing.csv. Spec 2.4 p.69.
type totals struct {
	zahlungen    dsfinvk.Decimal
	barzahlungen dsfinvk.Decimal
}

// aggregate holds the Kassenabschluss sums in first appearance order.
type aggregate struct {
	gvOrder  []gvKey
	gv       map[gvKey]amounts
	payOrder []payKey
	pay      map[payKey]dsfinvk.Decimal
	curOrder []string
	cur      map[string]dsfinvk.Decimal
	totals   totals
}

// aggregateClosing sums the Bons of type Beleg of one Kassenabschluss. Spec 2.4 p.31.
func aggregateClosing(c Kassenabschluss) (*aggregate, error) {
	a := &aggregate{
		gv:  make(map[gvKey]amounts),
		pay: make(map[payKey]dsfinvk.Decimal),
		cur: make(map[string]dsfinvk.Decimal),
	}

	for _, bon := range c.Bons {
		if bon.Typ != schema.BonTypBeleg {
			continue
		}
		if err := a.addPositions(bon); err != nil {
			return nil, fmt.Errorf("model: Bon %s: %w", bon.ID, err)
		}
		if err := a.addZahlungen(bon, c.Kasse.Basiswaehrung); err != nil {
			return nil, fmt.Errorf("model: Bon %s: %w", bon.ID, err)
		}
	}
	return a, nil
}

// addPositions sums the VAT shares of the Positionen of one Beleg. Spec 2.4 p.80, p.81.
func (a *aggregate) addPositions(bon Bon) error {
	for _, p := range bon.Positionen {
		for _, u := range p.USt {
			k := gvKey{typ: p.GVTyp, name: p.GVName, agentur: p.AgenturID, schluessel: u.Schluessel}
			if _, seen := a.gv[k]; !seen {
				a.gvOrder = append(a.gvOrder, k)
			}
			sum := a.gv[k]
			if err := sum.add(amounts{brutto: u.Brutto, netto: u.Netto, ust: u.USt}); err != nil {
				return err
			}
			a.gv[k] = sum
		}
	}
	return nil
}

// addZahlungen sums the payments of one Beleg. Spec 2.4 p.69, p.82, p.83.
func (a *aggregate) addZahlungen(bon Bon, base string) error {
	for _, z := range bon.Zahlungen {
		k := payKey{typ: z.Typ, name: z.Name}
		if _, seen := a.pay[k]; !seen {
			a.payOrder = append(a.payOrder, k)
		}
		sum, err := a.pay[k].Add(z.BetragBasis)
		if err != nil {
			return err
		}
		a.pay[k] = sum

		if a.totals.zahlungen, err = a.totals.zahlungen.Add(z.BetragBasis); err != nil {
			return err
		}
		if z.Typ != schema.ZahlartTypBar {
			continue
		}
		if a.totals.barzahlungen, err = a.totals.barzahlungen.Add(z.BetragBasis); err != nil {
			return err
		}
		if err := a.addCash(z, base); err != nil {
			return err
		}
	}
	return nil
}

// addCash sums the cash balance per currency. Spec 2.4 p.83.
func (a *aggregate) addCash(z Zahlung, base string) error {
	currency := currencyOf(z, base)
	if _, seen := a.cur[currency]; !seen {
		a.curOrder = append(a.curOrder, currency)
	}
	sum, err := a.cur[currency].Add(cashAmount(z, base))
	if err != nil {
		return err
	}
	a.cur[currency] = sum
	return nil
}

// currencyOf is the currency of a payment, defaulted to the Basiswaehrung. Spec 2.4 p.92.
func currencyOf(z Zahlung, base string) string {
	if z.Waehrung == "" {
		return base
	}
	return z.Waehrung
}

// cashAmount is the payment amount in its own currency. Spec 2.4 p.83, p.92.
func cashAmount(z Zahlung, base string) dsfinvk.Decimal {
	if currencyOf(z, base) == base {
		return z.BetragBasis
	}
	return z.BetragWaehrung
}
