--
-- PostgreSQL database dump
--

-- Dumped from database version 9.5.3
-- Dumped by pg_dump version 9.5.3

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

SET search_path = public, pg_catalog;

--
-- Data for Name: equity; Type: TABLE DATA; Schema: public; Owner: wcharczuk
--

COPY equity (id, active, name, ticker, exchange) FROM stdin;
1	t	SPDR S&P 500 ETF Trust	SPY	NYSEARCA
2	t	Vanguard MSCI EAFE ETF	VEA	NYSEARCA
3	t	Vanguard Total Stock Market ETF	VTI	NYSEARCA
4	t	Vanguard Value ETF	VTV	NYSEARCA
5	t	Vanguard Mid-Cap Value ETF	VOE	NYSEARCA
6	t	Vanguard Small-Cap Value ETF	VBR	NYSEARCA
7	t	Vanguard Emerging Markets Stock Index Fund	VWO	NYSEARCA
8	t	iShares National Muni Bond ETF	MUB	NYSEARCA
9	t	iShares iBoxx $ Investment Grade Corporate Bond ETF	LQD	NYSEARCA
10	t	Vanguard Total International Bond ETF	BNDX	NASDAQ
11	t	Vanguard Whitehall Funds	VWOB	NASDAQ
12	t	iPath S&P 500 VIX Short Term Futures TM ETN	VXX	NYSEARCA
13	t	Credit Suisse AG - VelocityShares Daily Inverse VIX Short 	VIX	NASDAQ
14	t	Credit Suisse AG - VelocityShares Daily 2x VIX Short Term	TVIX	NASDAQ
15	t	ProShares UltraShort S&P500 (ETF)	SDS	NYSEARCA
16	t	Microsoft Corporation	MSFT	NASDAQ
17	t	Tesla Motors Inc	TSLA	NASDAQ
18	t	Apple Inc.	AAPL	NASDAQ
19	t	Alphabet Inc.	GOOG	NASDAQ
20	t	Direxion Daily Gold Miners Bull 3X ETF	NUGT	NYSEARCA
21	t	SPDR Gold Trust (ETF)	GLD	NYSEARCA
22	t	VelocityShares 3X Long Crude ETN linked to the S&P GSCI 	UWTI	NYSEARCA
23	t	United States Oil Fund LP (ETF)	USO	NYSEARCA
24	t	VelocityShares 3X Inverse Crude ETN linked to the S&P	DWTI	NYSEARCA
25	t	ProShares UltraShort Bloomberg Crude Oil	SCO	NYSEARCA
26	t	Amazon.com, Inc.	AMZN	NASDAQ
27	t	Netflix, Inc.	NFLX	NASDAQ
28	t	Twitter Inc.	TWTR	NYSE
29	t	Facebook Inc.	FB	NASDAQ
30	t	LinkedIn Corp.	LNKD	NYSE
31	t	salesforce.com, inc.	CRM	NYSE
32	t	International Business Machines Corp.	IBM	NYSE
33	t	Infosys Ltd ADR	INFY	NYSE
34	t	SAP SE (ADR)	SAP	NYSE
35	t	Oracle Corporation	ORCL	NYSE
36	t	Intel Corporation	INTC	NASDAQ
37	t	NVIDIA Corporation	NVDA	NASDAQ
38	t	Advanced Micro Devices, Inc.	AMD	NASDAQ
39	t	ARM Holdings plc (ADR)	ARMH	NASDAQ
40	t	CurrencyShares British Pound Sterling Trust	FXB	NYSEARCA
41	t	CurrencyShares Euro Trust	FXE	NYSEARCA
42	t	Adobe Systems Incorporated	ADBE	NASDAQ
43	t	Autodesk, Inc.	ADSK	NASDAQ
44	t	Twilio Inc.	TWLO	NYSE
45	t	Cisco Systems, Inc.	CSCO	NASDAQ
46	t	Market Vectors Gold Miners ETF	GDX	NYSEARCA
47	t	iShares MSCI Emerging Markets Index ETF	EEM	NYSEARCA
48	t	ProShares Trust Ultra VIX Short Term Futures ETF	UVXY	NYSEARCA
49	t	Financial Select Sector SPDR Fund	XLF	NYSEARCA
50	t	iShares MSCI Japan ETF	EWJ	NYSEARCA
51	t	iShares Russell 2000 Index ETF	IWM	NYSEARCA
52	t	iShares MSCI EAFE Index Fund ETF	EFA	NYSEARCA
53	t	PowerShares QQQ Trust (Nasdaq-100) ETF	QQQ	NYSEARCA
54	t	iShares FTSE/Xinhua China 25 Index ETF	FXI	NYSEARCA
55	t	Direxion Daily Gold Miners Index Bear 3x Shares ETF	DUST	NYSEARCA
56	t	iShares MSCI Brazil Capped ETF	EWZ	NYSEARCA
57	t	SPDR S&P Oil & Gas Explore & Prod. ETF	XOP	NYSEARCA
58	t	Energy Select Sector SPDR (ETF)	XLE	NYSEARCA
59	t	iShares Silver Trust (ETF)	SLV	NYSEARCA
60	t	CurrencyShares Swedish Krona Trust	FXS	NYSEARCA
61	t	Exxon Mobil Corporation	XOM	NYSE
62	t	Berkshire Hathaway Inc.	BRK.A	NYSE
63	t	Berkshire Hathaway Inc.	BRK.B	NYSE
64	t	General Electric Company	GE	NYSE
65	t	Caterpillar Inc.	CAT	NYSE
66	t	Wells Fargo & Co	WFC	NYSE
67	t	AT&T Inc.	T	NYSE
68	t	Johnson & Johnson	JNJ	NYSE
69	t	JPMorgan Chase & Co.	JPM	NYSE
70	t	Wal-Mart Stores, Inc.	WMT	NYSE
71	t	Costco Wholesale Corporation	COST	NYSE
72	t	Target Corporation	TGT	NYSE
73	t	Procter & Gamble Co	PG	NYSE
74	t	Verizon Communications Inc.	VZ	NYSE
75	t	Pfizer Inc.	PFE	NYSE
76	t	Anheuser Busch Inbev SA (ADR)	BUD	NYSE
77	t	Alibaba Group Holding Ltd	BABA	NYSE
78	t	The Coca-Cola Co	KO	NYSE
79	t	Visa Inc	V	NYSE
80	t	American Express Company	AXP	NYSE
81	t	Home Depot Inc	HD	NYSE
82	t	Merck & Co., Inc.	MRK	NYSE
83	t	Comcast Corporation	CMCSA	NASDAQ
84	t	Walt Disney Co	DIS	NYSE
85	t	PepsiCo, Inc.	PEP	NYSE
86	t	CBS Corporation	CBS	NYSE
87	t	Citigroup Inc	C	NYSE
88	t	Goldman Sachs Group Inc	GS	NYSE
89	t	Morgan Stanley	MS	NYSE
90	t	Intuit Inc.	INTU	NASDAQ
91	t	Square Inc	SQ	NYSE
92	t	Market Vector Russia ETF Trust	RSX	NYSEARCA
93	t	Health Care SPDR (ETF)	XLV	NYSEARCA
94	t	Technology SPDR (ETF)	XLK	NYSEARCA
95	t	Utilities SPDR (ETF)	XLU	NYSEARCA
96	t	SPDR S&P Metals and Mining (ETF)	XME	NYSEARCA
97	t	SPDR S&P Retail (ETF)	XRT	NYSEARCA
98	t	SPDR S&P Pharmaceuticals (ETF)	XPH	NYSEARCA
100	t	SPDR S&P Biotech (ETF)	XBI	NYSEARCA
101	t	SPDR S&P Semiconductor (ETF)	XSD	NYSEARCA
102	t	SPDR S&P Oil & Gas Equipt & Servs. (ETF)	XES	NYSEARCA
103	t	SPDR S&P Capital Markets ETF	KCE	NYSEARCA
104	t	SPDR S&P Insurance ETF	KIE	NYSEARCA
105	t	iShares MSCI Germany Index Fund (ETF)	EWG	NYSEARCA
106	t	iShares MSCI Spain Capped ETF	EWP	NYSEARCA
107	t	iShares MSCI Canada Index (ETF)	EWC	NYSEARCA
108	t	iShares MSCI France Index (ETF)	EWQ	NYSEARCA
109	t	iShares MSCI South Korea Index Fund(ETF)	EWY	NYSEARCA
110	t	iShares MSCI Mexico Inv. Mt. Idx. (ETF)	EWW	NYSEARCA
111	t	iShares MSCI Italy Index (ETF)	EWI	NYSEARCA
112	t	iShares MSCI Australia Index Fund (ETF)	EWA	NYSEARCA
113	t	iShares MSCI Hong Kong Index Fund (ETF)	EWH	NYSEARCA
114	t	iShares MSCI South Africa Index (ETF)	EZA	NYSEARCA
115	t	ProShares UltraShort QQQ (ETF)	QID	NYSEARCA
\.


--
-- Name: equity_id_seq; Type: SEQUENCE SET; Schema: public; Owner: wcharczuk
--

SELECT pg_catalog.setval('equity_id_seq', 117, true);


--
-- PostgreSQL database dump complete
--

