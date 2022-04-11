import React, { useContext, useState } from 'react'

import { WebSocketContext } from '../WebSocket'
///////////////////// metamask s ////////////////
import {useWeb3React} from '@web3-react/core'
import {injected} from './lib/connectors'

const Join = () => {
  const ws = useContext(WebSocketContext)
  const [isJoining, setJoining] = useState(false)
  const [username, setUsername] = useState('')
  const [roomId, setRoomid] = useState('')
  ///////////////////// metamask s ////////////////
  const { chainId, account, active, activate, deactivate } = useWeb3React()
  const handleConnect = () => {
    if(active) { 
      // document.getElementById('user_address').value = {chainId}
      deactivate()
      return
    }
    activate(injected,(error)=>{
      if('/No Ethereum provider was found on window.ethereum/'.test(error)){
        window.open('https://metamask.io/download.html')
      }
    })
  }
  ///////////////////// metamask e ////////////////
  
  ////////////////////////////////////////////////////////////////////////
  const current = decodeURI(window.location.href)
  const search = current.split("?")[1]
  const params = new URLSearchParams(search)
  // const c4ei_Address = params.get('c4ei_Address')
  let reqroomid = params.get('room')
  // const { body: { c4ei_Address, sec2 }, } = req
  const hstyle = { display: "block" /* display: "block" */ }

  const handleJoinGame = (username) => {
    const trimmedUsername = username.trim()
    if (trimmedUsername) {
      setJoining(true)
      if(reqroomid==""||reqroomid==null)
      {
        reqroomid=1
      } 
      setRoomid(reqroomid)
      console.log(roomId+":roomId")
      ws.joinGame(username)
      // ws.joinGame(username , roomId) //220411
      // ws.joinGame2(roomId, username)
    }
  }

  const handleKeyPress = (e) => {
    if (e.code === 'Enter') {
      handleJoinGame(e.target.value)
    }
  }

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
            <button type="button" onClick={handleConnect} className="border rounded-sm mr-2 p-3">
              {active?'disconnect':'connect'}
            </button>
          </div>

          <div>
            <p>Account: {account}</p>
            <p>ChainId: {chainId}</p>
          </div>

          <label className="sr-only" htmlFor="name">Name</label>
          <input
            autoFocus
            className="border rounded-sm mr-2 p-3"
            id="user_address"
            name="name"
            onChange={(e) => setUsername(e.target.value)}
            onKeyUp={handleKeyPress}
            placeholder="c4ei_Address"
            type="text"
            value={account}
            // value={username}
            // readOnly
          />
          <button id="btnGo"
            className="bg-blue-700 px-6 py-3 font-bold text-white"
            disabled={isJoining}
            onClick={() => handleJoinGame(username)}
          >
            Join
          </button>
	      </div>
      </div>
    </div>
  )
}

export default Join
