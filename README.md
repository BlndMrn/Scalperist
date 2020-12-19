# Scalperist
Language: [Ru](https://github.com/BlndMrn/Scalperist/blob/main/readme/readme.ru.md)  
[Trades example](https://www.youtube.com/watch?v=ys-YOsoCF34)  
Bot trades on high volatility markets.  Price change on more than 0.1% is trigger for making orders against price move.
- Take profit: 0.12%.
- Leverage: Cross.
- Stoploss: 10%.
- Percent of balance on first order 3%, each next order quantity +3% to first order quantity. 

### Problems
When volatility is extremly high encouter problems with orders creation. Looks like Bybit exchange can't handle so many trades at once.  
Searching for solutions.

##### Imports
- github.com/frankrap/bybit-api/rest

### Donations
BTC: 1DXBapnm5H8djZBEBd1LKSYpy6hRepBr7u  
ETH ERC20: 0x8a5df2a88a6f7c4116d6ff590a0eb2102cdae5b6  
USDT TRC20: TEWZoBgyLDRHWqKCCEjSkqRW1WnB3WL2hR  
