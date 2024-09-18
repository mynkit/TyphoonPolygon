# TyphoonPolygon

## How to Run

### Install

Golang (>= 1.21)

```sh
# Install GEOS
# https://libgeos.org/usage/install/
brew install geos
```

```sh
go mod download
```

Python (>= 3.10)

```sh
pip install beautifulsoup4 lxml
```

### Run

```sh
python parse_xmls.py # jsonディレクトリにxmlのscraping結果が書き出される
```

```sh
go run main.go # output.geojsonが書き出される
```

## Show GeoJSON

https://geojson.io/

に`output.geojson`の結果を貼り付ければGeoJSONの確認が可能