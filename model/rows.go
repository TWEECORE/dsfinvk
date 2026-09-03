package model

import (
	"github.com/tweecore/dsfinvk"
	"github.com/tweecore/dsfinvk/schema"
)

// closingRow writes the cashpointclosing.csv row. Spec 2.4 p.66 to p.69.
func (b *builder) closingRow(c Kassenabschluss, sums totals) {
	r := b.rec("cashpointclosing.csv")
	r.day("Z_BUCHUNGSTAG", c.Buchungstag)
	r.enum("TAXONOMIE_VERSION", taxonomieVersion(c), schema.CurrentTaxonomyVersions())
	r.text("Z_START_ID", startID(c.Bons))
	r.text("Z_ENDE_ID", endID(c.Bons))
	r.text("NAME", c.Unternehmen.Name)
	r.text("STRASSE", c.Unternehmen.Strasse)
	r.text("PLZ", c.Unternehmen.PLZ)
	r.text("ORT", c.Unternehmen.Ort)
	r.text("LAND", c.Unternehmen.Land)
	r.text("STNR", c.Unternehmen.StNr)
	r.text("USTID", c.Unternehmen.UStID)
	r.num("Z_SE_ZAHLUNGEN", sums.zahlungen)
	r.num("Z_SE_BARZAHLUNGEN", sums.barzahlungen)
	b.add(r)
}

// taxonomieVersion returns the TAXONOMIE_VERSION of c, defaulted. Spec 2.4 p.66.
func taxonomieVersion(c Kassenabschluss) string {
	if c.TaxonomieVersion == "" {
		return defaultTaxonomieVersion
	}
	return c.TaxonomieVersion
}

// startID is the BON_ID of the first Vorgang of the closing. Spec 2.4 p.66.
func startID(bons []Bon) string {
	if len(bons) == 0 {
		return ""
	}
	return bons[0].ID
}

// endID is the BON_ID of the last Vorgang of the closing. Spec 2.4 p.67.
func endID(bons []Bon) string {
	if len(bons) == 0 {
		return ""
	}
	return bons[len(bons)-1].ID
}

// locationRow writes the location.csv row. Spec 2.4 p.69, p.70.
func (b *builder) locationRow(l Location) {
	r := b.rec("location.csv")
	r.text("LOC_NAME", l.Name)
	r.text("LOC_STRASSE", l.Strasse)
	r.text("LOC_PLZ", l.PLZ)
	r.text("LOC_ORT", l.Ort)
	r.text("LOC_LAND", l.Land)
	r.text("LOC_USTID", l.UStID)
	b.add(r)
}

// cashregisterRow writes the cashregister.csv row. Spec 2.4 p.71, p.72.
func (b *builder) cashregisterRow(k Kasse) {
	r := b.rec("cashregister.csv")
	r.text("KASSE_BRAND", k.Brand)
	r.text("KASSE_MODELL", k.Modell)
	r.text("KASSE_SERIENNR", k.Seriennr)
	r.text("KASSE_SW_BRAND", k.SWBrand)
	r.text("KASSE_SW_VERSION", k.SWVersion)
	r.text("KASSE_BASISWAEH_CODE", k.Basiswaehrung)
	r.flag("KEINE_UST_ZUORDNUNG", k.KeineUStZuordnung)
	b.add(r)
}

// terminalRow writes one slaves.csv row. Spec 2.4 p.73, p.74.
func (b *builder) terminalRow(t Terminal) {
	r := b.rec("slaves.csv")
	r.text("TERMINAL_ID", t.ID)
	r.text("TERMINAL_BRAND", t.Brand)
	r.text("TERMINAL_MODELL", t.Modell)
	r.text("TERMINAL_SERIENNR", t.Seriennr)
	r.text("TERMINAL_SW_BRAND", t.SWBrand)
	r.text("TERMINAL_SW_VERSION", t.SWVersion)
	b.add(r)
}

// agenturRow writes one pa.csv row. Spec 2.4 p.74 to p.76.
func (b *builder) agenturRow(a Agentur) {
	r := b.rec("pa.csv")
	r.id("AGENTUR_ID", a.ID)
	r.text("AGENTUR_NAME", a.Name)
	r.text("AGENTUR_STRASSE", a.Strasse)
	r.text("AGENTUR_PLZ", a.PLZ)
	r.text("AGENTUR_ORT", a.Ort)
	r.text("AGENTUR_LAND", a.Land)
	r.text("AGENTUR_STNR", a.StNr)
	r.text("AGENTUR_USTID", a.UStID)
	b.add(r)
}

// tseRow writes one tse.csv row, splitting the certificate. Spec 2.4 p.77 to p.79.
func (b *builder) tseRow(t TSE) {
	r := b.rec("tse.csv")
	r.id("TSE_ID", t.ID)
	r.text("TSE_SERIAL", t.Serial)
	r.enum("TSE_SIG_ALGO", string(t.SigAlgo), schema.TSESigAlgoValues())
	r.enum("TSE_ZEITFORMAT", string(t.Zeitformat), schema.TSEZeitformatValues())
	r.enum("TSE_PD_ENCODING", string(t.PDEncoding), schema.TSEPDEncodingValues())
	r.text("TSE_PUBLIC_KEY", t.PublicKey)
	for i, chunk := range certChunks(t.Zertifikat) {
		r.text(certColumn(i+1), chunk)
	}
	b.add(r)
}

// vatRow writes one vat.csv row. Spec 2.4 p.76, p.77.
func (b *builder) vatRow(u USt) {
	r := b.rec("vat.csv")
	r.id("UST_SCHLUESSEL", u.Schluessel)
	r.num("UST_SATZ", u.Satz)
	r.text("UST_BESCHR", u.Beschreibung)
	b.add(r)
}

// businesscaseRow writes one businesscases.csv row. Spec 2.4 p.80, p.81.
func (b *builder) businesscaseRow(k gvKey, v amounts) {
	r := b.rec("businesscases.csv")
	r.text("GV_TYP", string(k.typ))
	r.text("GV_NAME", k.name)
	r.id("AGENTUR_ID", k.agentur)
	r.id("UST_SCHLUESSEL", k.schluessel)
	r.num("Z_UMS_BRUTTO", v.brutto)
	r.num("Z_UMS_NETTO", v.netto)
	r.num("Z_UST", v.ust)
	b.add(r)
}

// paymentRow writes one payment.csv row. Spec 2.4 p.82.
func (b *builder) paymentRow(k payKey, v dsfinvk.Decimal) {
	r := b.rec("payment.csv")
	r.text("ZAHLART_TYP", string(k.typ))
	r.text("ZAHLART_NAME", k.name)
	r.num("Z_ZAHLART_BETRAG", v)
	b.add(r)
}

// currencyRow writes one cash_per_currency.csv row. Spec 2.4 p.83.
func (b *builder) currencyRow(currency string, v dsfinvk.Decimal) {
	r := b.rec("cash_per_currency.csv")
	r.text("ZAHLART_WAEH", currency)
	r.num("ZAHLART_BETRAG_WAEH", v)
	b.add(r)
}

// bonRow writes the transactions.csv row. Spec 2.4 p.84 to p.89.
func (b *builder) bonRow(bon Bon) {
	r := b.rec("transactions.csv")
	r.text("BON_ID", bon.ID)
	r.id("BON_NR", bon.Nr)
	r.enum("BON_TYP", string(bon.Typ), schema.BonTypValues())
	r.text("BON_NAME", bon.Name)
	r.text("TERMINAL_ID", bon.TerminalID)
	r.flag("BON_STORNO", bon.Storno)
	r.stamp("BON_START", bon.Start)
	r.stamp("BON_ENDE", bon.Ende)
	r.text("BEDIENER_ID", bon.BedienerID)
	r.text("BEDIENER_NAME", bon.BedienerName)
	r.num("UMS_BRUTTO", bon.UmsBrutto)
	r.text("KUNDE_NAME", bon.Kunde.Name)
	r.text("KUNDE_ID", bon.Kunde.ID)
	r.text("KUNDE_TYP", bon.Kunde.Typ)
	r.text("KUNDE_STRASSE", bon.Kunde.Strasse)
	r.text("KUNDE_PLZ", bon.Kunde.PLZ)
	r.text("KUNDE_ORT", bon.Kunde.Ort)
	r.text("KUNDE_LAND", bon.Kunde.Land)
	r.text("KUNDE_USTID", bon.Kunde.UStID)
	r.text("BON_NOTIZ", bon.Notiz)
	b.add(r)
}

// allocationRow writes one allocation_groups.csv row. Spec 2.4 p.89.
func (b *builder) allocationRow(bonID, kreis string) {
	r := b.rec("allocation_groups.csv")
	r.text("BON_ID", bonID)
	r.text("ABRECHNUNGSKREIS", kreis)
	b.add(r)
}

// zahlungRow writes one datapayment.csv row. Spec 2.4 p.91, p.92.
func (b *builder) zahlungRow(bonID string, z Zahlung) {
	currency := currencyOf(z, b.basiswaehrung)

	r := b.rec("datapayment.csv")
	r.text("BON_ID", bonID)
	r.enum("ZAHLART_TYP", string(z.Typ), schema.ZahlartTypValues())
	r.text("ZAHLART_NAME", z.Name)
	r.text("ZAHLWAEH_CODE", currency)
	if currency != b.basiswaehrung {
		r.num("ZAHLWAEH_BETRAG", z.BetragWaehrung)
	}
	r.num("BASISWAEH_BETRAG", z.BetragBasis)
	b.add(r)
}

// bonUStRow writes one transactions_vat.csv row of printed totals. Spec 2.4 p.90, p.91.
func (b *builder) bonUStRow(bonID string, u BonUSt) {
	b.checkUSt(u.Schluessel)

	r := b.rec("transactions_vat.csv")
	r.text("BON_ID", bonID)
	r.id("UST_SCHLUESSEL", u.Schluessel)
	r.num("BON_BRUTTO", u.Brutto)
	r.num("BON_NETTO", u.Netto)
	r.num("BON_UST", u.USt)
	b.add(r)
}

// referenzRow writes one references.csv row. Spec 2.4 p.102, p.103.
func (b *builder) referenzRow(bonID string, ref Referenz) {
	r := b.rec("references.csv")
	r.text("BON_ID", bonID)
	r.text("POS_ZEILE", ref.PosZeile)
	r.enum("REF_TYP", string(ref.Typ), schema.RefTypValues())
	r.text("REF_NAME", ref.Name)
	if ref.Typ == schema.RefTypTransaktion {
		r.stamp("REF_DATUM", ref.Datum)
		r.text("REF_Z_KASSE_ID", ref.KasseID)
		if ref.Nr != 0 {
			r.id("REF_Z_NR", ref.Nr)
		}
	}
	r.text("REF_BON_ID", ref.BonID)
	b.add(r)
}

// tseTransaktionRow writes one transactions_tse.csv row. Spec 2.4 p.104, p.105.
func (b *builder) tseTransaktionRow(bonID string, t TSETransaktion) {
	b.checkTSE(t.TSEID)

	r := b.rec("transactions_tse.csv")
	r.text("BON_ID", bonID)
	r.id("TSE_ID", t.TSEID)
	r.id("TSE_TANR", t.Nr)
	r.tseStamp("TSE_TA_START", t.Start)
	r.tseStamp("TSE_TA_ENDE", t.Ende)
	r.enum("TSE_TA_VORGANGSART", string(t.Vorgangsart), schema.TSEVorgangsartValues())
	r.id("TSE_TA_SIGZ", t.SigZaehler)
	r.text("TSE_TA_SIG", t.Signatur)
	r.text("TSE_TA_FEHLER", t.Fehler)
	r.text("TSE_VORGANGSDATEN", t.Vorgangsdaten)
	b.add(r)
}

// positionRow writes one lines.csv row. Spec 2.4 p.93 to p.96.
func (b *builder) positionRow(bonID string, p Position) {
	b.checkAgentur(p.AgenturID)

	r := b.rec("lines.csv")
	r.text("BON_ID", bonID)
	r.text("POS_ZEILE", p.Zeile)
	r.text("GUTSCHEIN_NR", p.GutscheinNr)
	r.text("ARTIKELTEXT", p.Artikeltext)
	r.text("POS_TERMINAL_ID", p.TerminalID)
	r.enum("GV_TYP", string(p.GVTyp), schema.GVTypValues())
	r.text("GV_NAME", p.GVName)
	r.flag("INHAUS", p.Inhaus)
	r.flag("P_STORNO", p.Storno)
	r.id("AGENTUR_ID", p.AgenturID)
	r.text("ART_NR", p.ArtNr)
	r.text("GTIN", p.GTIN)
	r.text("WARENGR_ID", p.WarengrID)
	r.text("WARENGR", p.Warengr)
	r.numOpt("MENGE", p.Menge)
	r.numOpt("FAKTOR", p.Faktor)
	r.text("EINHEIT", p.Einheit)
	r.numOpt("STK_BR", p.StkBrutto)
	b.add(r)
}

// posUStRow writes one lines_vat.csv row. Spec 2.4 p.97.
func (b *builder) posUStRow(bonID, zeile string, u PosUSt) {
	b.checkUSt(u.Schluessel)

	r := b.rec("lines_vat.csv")
	r.text("BON_ID", bonID)
	r.text("POS_ZEILE", zeile)
	r.id("UST_SCHLUESSEL", u.Schluessel)
	r.numOpt("POS_BRUTTO", u.Brutto)
	r.numOpt("POS_NETTO", u.Netto)
	r.numOpt("POS_UST", u.USt)
	b.add(r)
}

// preisfindungRow writes one itemamounts.csv row. Spec 2.4 p.98, p.99.
func (b *builder) preisfindungRow(bonID, zeile string, p Preisfindung) {
	b.checkUSt(p.Schluessel)

	r := b.rec("itemamounts.csv")
	r.text("BON_ID", bonID)
	r.text("POS_ZEILE", zeile)
	r.enum("TYP", string(p.Typ), schema.PreisfindungTypValues())
	r.id("UST_SCHLUESSEL", p.Schluessel)
	r.numOpt("PF_BRUTTO", p.Brutto)
	r.numOpt("PF_NETTO", p.Netto)
	r.numOpt("PF_UST", p.USt)
	b.add(r)
}

// zusatzinfoRow writes one subitems.csv row. Spec 2.4 p.99 to p.101.
func (b *builder) zusatzinfoRow(bonID, zeile string, z Zusatzinfo) {
	b.checkUSt(z.Schluessel)

	r := b.rec("subitems.csv")
	r.text("BON_ID", bonID)
	r.text("POS_ZEILE", zeile)
	r.text("ZI_ART_NR", z.ArtNr)
	r.text("ZI_GTIN", z.GTIN)
	r.text("ZI_NAME", z.Name)
	r.text("ZI_WARENGR_ID", z.WarengrID)
	r.text("ZI_WARENGR", z.Warengr)
	r.num("ZI_MENGE", z.Menge)
	r.numOpt("ZI_FAKTOR", z.Faktor)
	r.text("ZI_EINHEIT", z.Einheit)
	r.id("ZI_UST_SCHLUESSEL", z.Schluessel)
	r.numOpt("ZI_BASISPREIS_BRUTTO", z.BasispreisBrutto)
	r.numOpt("ZI_BASISPREIS_NETTO", z.BasispreisNetto)
	r.numOpt("ZI_BASISPREIS_UST", z.BasispreisUSt)
	b.add(r)
}
