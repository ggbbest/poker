# Poker WebRTC

Experimenting with WebRTC by making a poker game.

https://richard.to/programming/poker-game-part-1-intro.html

go 모른다 ㅠㅠ 그러나 한번 덤벼보자 21-11-12

cd /cmd/poker-app
go run main.go

/home/dev/www/c4ei.net/poker/client/poker-app
sudo npm install

Cannot find module 'tailwindcss'
npm i tailwindcss

yarn start       
yarn run v1.22.15
$ craco start
(node:11884) UnhandledPromiseRejectionWarning: /home/dev/www/c4ei.net/poker/client/poker-app/node_modules/color/index.js:257
			lum[i] = (chan <= 0.039_28) ? chan / 12.92 : ((chan + 0.055) / 1.055) ** 2.4;
			                  ^^^^^
node -v
v10.19.0
sudo npm install -g n
sudo n stable
PATH="$PATH"

nvm install 16.13.0
cd /home/dev/www/vpoker.c4ei.net/client/poker-app
yarn start
###################################################
cd /home/dev/www/vpoker.c4ei.net/cmd/poker-app
go run main.go
###################################################

/home/dev/www/vpoker.c4ei.net/client/poker-app/src/routes/Join.js

/home/dev/www/vpoker.c4ei.net/cmd/poker-app/main.go
go get github.com/go-sql-driver/mysql 
go get github.com/joho/godotenv


/home/dev/www/vpoker.c4ei.net/pkg/server/.env
/home/dev/www/vpoker.c4ei.net/pkg/server/game.go  --> func dbcall() {

/home/dev/www/vpoker.c4ei.net/client/poker-app/src/routes/Join.js
c4ei_Address login 

/home/dev/www/vpoker.c4ei.net/pkg/server/game.go
	// Link user with player seat
	selectedPlayer.Name = c.username
	selectedPlayer.Chips = defaultChips    --> 
	selectedPlayer.Status = poker.PlayerSittingOut
	selectedPlayer.IsHuman = true
	c.seatID = selectedPlayer.ID


// 1. int to string - 숫자(정수)를 문자열로 변환
a := strconv.Itoa(100)
fmt.Println("a: ", a)                      // a: 100
fmt.Println("type a: ", reflect.TypeOf(a)) // type a: string

// 1-1. int to string - 100을 10진수 문자열로 변환
aa := strconv.FormatInt(100, 10)
fmt.Println("aa: ", aa)                      // aa: 100
fmt.Println("type aa: ", reflect.TypeOf(aa)) // type aa: string

// 2. string to int - 문자열을 숫자(정수) 변환
b, _ := strconv.Atoi("100")
fmt.Println("b: ", b)                      // b:  100
fmt.Println("type b: ", reflect.TypeOf(b)) // type b: int

bb, _ := strconv.ParseInt("100", 10, 64)
fmt.Println("bb: ", bb)                      // bb: 100
fmt.Println("type bb: ", reflect.TypeOf(bb)) // type bb: int64

ℝ --> 🍺

https://github.com/ggbbest/poker
room param add - 220306


--
const joinGame = (client, username, roomid ) => {