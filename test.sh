echo '\ncurl -i -X GET http://localhost:1576/align/\n\n'
curl -i -X GET http://localhost:1576/align
echo '\ncurl -i -X GET http://localhost:1576/align?area=Science&year=6,7,8\n\n'
curl -i -X GET "http://localhost:1576/align?area=Science&year=6,7,8"
echo '\ncurl -i -X GET http://localhost:1576/align?year=6,7,8\n\n'
curl -i -X GET "http://localhost:1576/align?year=6,7,8"
echo '\ncurl -i -X GET http://localhost:1576/align?year=6,7,8&item=SC4_12ES,SC4_13ES\n\n'
curl -i -X GET "http://localhost:1576/align?year=6,7,8&item=SC4_12ES,SC4_13ES"
echo '\ncurl -i -X GET http://localhost:1576/index?search=Biotechnology\n\n'
curl -i -X GET http://localhost:1576/index?search=Biotechnology
echo '\ncurl -i -X GET http://localhost:1576/index?search=Claudius\n\n'
curl -i -X GET http://localhost:1576/index?search=Claudius
