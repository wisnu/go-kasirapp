package models

type ProdukTerlaris struct {
	Nama       string `json:"nama"`
	QtyTerjual int    `json:"qty_terjual"`
}

type DailyReport struct {
	TotalRevenue    int            `json:"total_revenue"`
	TotalTransaksi  int            `json:"total_transaksi"`
	ProdukTerlaris  ProdukTerlaris `json:"produk_terlaris"`
}
