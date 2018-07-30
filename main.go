package main

import (
	"fmt"
	"github.com/adshao/go-binance"
	"context"
	"strings"
	"strconv"
)

var (
	apiKey = "XX"
	secretKey = "XX"
	Threshold = 1.2 //利润
	Fee = 0.6  //手续费
	Tradesize = 0.0015
)
var client = binance.NewClient(apiKey, secretKey)

func minqrt() (map[string]string) {
	var symbolinfos map[string]string
	symbolinfos = make(map[string]string)
	info,err := client.NewExchangeInfoService().Do(context.Background())
	if err != nil{
		return nil
	}
	for _,each:= range info.Symbols{
		symbolinfos[each.Symbol] = each.Filters[1]["minQty"]
	}
	return symbolinfos
}

func forward_check_profit_coin (coins []string, prices map[string]float64, threshold float64, fee float64) (Coin string, Profit float64)  {
	var highProfit float64 = 0.0
	var highProfitcoin string = ""
	for _,coin := range coins{
		basesymbol := coin+"BTC"
		midsymbol := coin+"ETH"
		quotesymbol := "ETHBTC"
		baseprice := prices[basesymbol]
		midprice := prices[midsymbol]
		quoteprice := prices[quotesymbol]
		if (baseprice != 0)&&(midprice != 0)&&(quoteprice != 0) {
			arbprice := midprice*quoteprice
			diff := ((arbprice/baseprice) - 1) * 100.0
			profit := diff - fee
			if profit > threshold {
				if profit > highProfit {
					highProfit = profit
					highProfitcoin = coin
				}
			}
		}
	}
	return highProfitcoin, highProfit
}

func buy_qty(symbol string, symbolinfo map[string]string,qty float64) (res_qty float64) {

	minQty,_ := strconv.ParseFloat(symbolinfo[symbol],64)
	if qty < minQty{
		qty = minQty
	} else {
		temp := int64(qty*100000000)%int64(minQty*100000000)
		ftemp := float64(temp)/100000000
		qty = qty-ftemp
	}
	return qty
}

func get_balance(coin string) {
	res, err := client.NewGetAccountService().Do(context.Background())
	//var count float64
	if err != nil {
		fmt.Println(err)
		return
	}
	for _,each := range res.Balances{
		if each.Asset == coin{

			fmt.Println(each.Free)
			//return
		}
	}
}

func sell_qty(coin string,symbol string,symbolinfo map[string]string) (qty float64){
	res, err := client.NewGetAccountService().Do(context.Background())
	var count float64
	if err != nil {
		fmt.Println(err)
		return
	}
	for _,each := range res.Balances{
		if each.Asset == coin{
			count,_ = strconv.ParseFloat(each.Free,64)
		}
	}

	minQty,_ := strconv.ParseFloat(symbolinfo[symbol],64)
	if count < minQty{
		count = minQty
	} else {
		temp := int64(count*100000000)%int64(minQty*100000000)
		ftemp := float64(temp)/100000000
		count = count-ftemp
	}
	return count
}


func forward_excecute(coin string, btcvalue float64, prices map[string]float64,symbolinfo map[string]string ) {
	btcsymbol := coin+"BTC"
	btcsymbolprice := prices[btcsymbol]
	qty :=  btcvalue/btcsymbolprice
	btcqty := buy_qty(btcsymbol,symbolinfo,qty)
	buyvalue := btcqty * btcsymbolprice
	if buyvalue < btcvalue * 1.1{
		qtystr := strconv.FormatFloat(btcqty,'g',-1,64)
		//fmt.Println("btcbuy",qty,btcqty,qtystr,coin)
		_, err1 := client.NewCreateOrderService().Symbol(btcsymbol).Side(binance.SideTypeBuy).Type(binance.OrderTypeMarket).Quantity(qtystr).Do(context.Background())
		if err1 != nil {
			fmt.Println(err1)
			return
		}
		midsymbol := coin+"ETH"
		ethqty := sell_qty(coin,midsymbol,symbolinfo)
		//fmt.Println(ethqty)
		ethqtystr := strconv.FormatFloat(ethqty,'g',-1,64)
		_, err2 := client.NewCreateOrderService().Symbol(midsymbol).Side(binance.SideTypeSell).Type(binance.OrderTypeMarket).Quantity(ethqtystr).Do(context.Background())
		if err2 != nil {
			fmt.Println(err2)
			return
		}
		quotesymbol := "ETHBTC"
		ethbtcqty := sell_qty("ETH",quotesymbol,symbolinfo)
		//fmt.Println(ethbtcqty)
		ethbtcqtystr := strconv.FormatFloat(ethbtcqty,'g',-1,64)
		_, err3 := client.NewCreateOrderService().Symbol(quotesymbol).Side(binance.SideTypeSell).Type(binance.OrderTypeMarket).Quantity(ethbtcqtystr).Do(context.Background())
		if err3 != nil {
			fmt.Println(err2)
			return
		}
		get_balance("BTC")
	}
}





func arb(symbolinfo map[string]string)  {
	var COINS [] string
	var prices_map map[string]float64
	prices_map = make(map[string]float64)
	prices, err := client.NewListPricesService().Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, pair := range prices {
		prices_map[pair.Symbol],_= strconv.ParseFloat(pair.Price,64)
		coin := pair.Symbol
		coin  = strings.Replace(coin,"BTC","",-1)
		coin  = strings.Replace(coin,"ETH","",-1)
		coin  = strings.Replace(coin,"BNB","",-1)
		COINS = append(COINS, coin)
		}
		 Coin,Profit := forward_check_profit_coin(COINS,prices_map,Threshold,Fee)
		 if Profit > 0{
			 forward_excecute(Coin,Tradesize,prices_map,symbolinfo)
		 }
}



func main() {

	symbolinfo := minqrt()
	for{
		arb(symbolinfo)
	}


}
