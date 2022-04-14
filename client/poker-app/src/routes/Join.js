import React, { useContext, useState } from 'react'

import { WebSocketContext } from '../WebSocket'
///////////////////// metamask s ////////////////
import {useWeb3React} from '@web3-react/core'
import {injected} from './lib/connectors'

const Join = () => {
  const ws = useContext(WebSocketContext)
  const [isJoining, setJoining] = useState(false)
  // const [username, setUsername] = useState('')
  // const [roomId, setRoomid] = useState('')
  ///////////////////// metamask s ////////////////
  const { chainId, account, active, activate, deactivate } = useWeb3React()
  const handleConnect = () => {
    if(active) { 
      deactivate()
      handleJoinGame(account) // join game
      return
    }
    activate(injected,(error)=>{
      if('No Ethereum provider was found on window.ethereum'.test(error)){
        window.open('https://metamask.io/download.html')
      }
    })
  }
  ///////////////////// metamask e ////////////////
  
  ////////////////////////////////////////////////////////////////////////
  // const current = decodeURI(window.location.href)
  // const search = current.split("?")[1]
  // const params = new URLSearchParams(search)
  // let reqroomid = params.get('room')
  // const { body: { c4ei_Address, sec2 }, } = req
  const hstyle = { display: "block" /* display: "block" */ }

  const handleJoinGame = (account) => {
    // console.log("##### /src/routes/Join.js ##### : handleJoinGame 1")
    if(account===undefined||account===""||account===null){
      alert("Metamask 로그인 후 접속 부탁드립니다.")
      return
    }
    const trimmedUsername = account.trim()
    // console.log("##### /src/routes/Join.js ##### : handleJoinGame 2")
    if (trimmedUsername) {
      setJoining(true)
      // console.log("##### /src/routes/Join.js ##### : handleJoinGame 3")
      // if(reqroomid==""||reqroomid==null) { reqroomid=1 } 
      // setRoomid(reqroomid)
      // console.log("##### /src/routes/Join.js #####"+roomId+":roomId")
      // console.log("##### /src/routes/Join.js ##### : handleJoinGame 4")
      ws.joinGame(account)
      // ws.joinGame(username , roomId) //220411
      // ws.joinGame2(roomId, username)
    }
  }

  // const handleKeyPress = (e) => {
  //   if (e.code === 'Enter') {
  //     // handleJoinGame(e.target.value)
  //     handleJoinGame(account)
  //   }
  // }

  // window.onload = function() { setTimeout(() => { jsfn_btnGo() }, 900) }
  // function jsfn_btnGo(){
  //   // alert(current + "/"+search + "/"+params + "/"+c4ei_Address)
  //   // alert(c4ei_Address)

  //   if(c4ei_Address==""||c4ei_Address==null){
  //     // let _email = req.body.c4ei_Address;
  //     // if(_email==""||_email==null){
  //       alert("c4ei.net 에서 로그인 후 접속 부탁드립니다.")
  //       document.location.href="https://c4ei.net"
  //     // }
  //   }else{
  //     handleJoinGame(c4ei_Address)
  //   }
  // }
  
////////////////////////////////////////////////////////////////////////
  return (

    <div className="container mx-auto" style={hstyle}>
      <div className="flex h-screen">
        <div className="border shadow-lg m-auto p-10 font-bold text-black">
          <div>
            <button type="button" onClick={handleConnect} className="bg-blue-700 px-6 py-3 font-bold text-white">
              {active?'disconnect':'①connect'}
            </button>
            <br/>
            <p>Account: {account}</p>
            <p>ChainId: {chainId}</p>
          </div>

          <label className="sr-only" htmlFor="name">Name</label>
          <input
            // autoFocus
            className="border rounded-sm mr-2 p-3"
            name="name"
            // onChange={(e) => setUsername(e.target.value)}
            // onChange={() => handleJoinGame(account)}
            // onKeyUp={handleKeyPress}
            placeholder="login metamask button click first!"
            type="text"
            value={account}
            // value={username}
            readOnly = {true}
          />
          <button id="btnGo"
            className="bg-blue-700 px-6 py-3 font-bold text-white"
            disabled={isJoining}
            onClick={() => handleJoinGame(account)}
          >
            ②Join
          </button>
	      </div>
      </div>
    </div>
  )
}

export default Join
