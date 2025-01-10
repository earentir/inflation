# good source for EU rates
https://data.ecb.europa.eu/data/data-categories/macroeconomic-and-sectoral-statistics/consumer-prices-and-inflation/total?searchTerm=&filterSequence=.frequency.reference_area_name&sort=relevance&pageSize=10&filterType=basic&showDatasetModal=false&filtersReset=false&resetAll=false&reference_area_name%5B0%5D=Germany&frequency%5B0%5D=M&resetAllFilters=false&tags_array%5B0%5D=Overall+index&tags_array%5B1%5D=Financial+market


# check the cost change for US from 2003 to 2024 for 35 USD
./inflationcmd --inflation-list https://raw.githubusercontent.com/earentir/inflation/refs/heads/main/data/inflationratelist.json compare US 2001 2024 35

# Average inflation rate for CH for 2024
./inflationcmd --inflation-list ../data/inflationratelist.json year CH 2024

# List Countries in data
./inflationcmd --inflation-list ../data/inflationratelist.json listCountries
