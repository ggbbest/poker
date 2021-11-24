import React, { useContext, useState } from 'react'

import { WebSocketContext } from '../WebSocket'
const Join = () => {
  const ws = useContext(WebSocketContext)

  const [isJoining, setJoining] = useState(false)
  const [username, setUsername] = useState('')

  const handleJoinGame = (username) => {
    const trimmedUsername = username.trim()
    if (trimmedUsername) {
      setJoining(true)
      ws.joinGame(username)
    }
  }

  const handleKeyPress = (e) => {
    if (e.code === 'Enter') {
      handleJoinGame(e.target.value)
    }
  }
////////////////////////////////////////////////////////////////////////
const current = decodeURI(window.location.href)
const search = current.split("?")[1]
const params = new URLSearchParams(search)
const email = params.get('email')

const hstyle = {
  display: "none"
  // display: "block"
}

window.onload = function() {
  setTimeout(() => { jsfn_btnGo() }, 900)
}

function jsfn_btnGo(){
  // alert(current + "/"+search + "/"+params + "/"+email)
  // alert(email)
  if(email==""||email==null){
    alert("c4ei.net 에서 로그인 후 접속 부탁드립니다.")
    document.location.href="https://c4ei.net"
  }else{
    handleJoinGame(email)
  }
}
////////////////////////////////////////////////////////////////////////
  return (

    <div className="container mx-auto" style={hstyle}>
      <div className="flex h-screen">
        <div className="border shadow-lg m-auto p-10 font-bold text-black">
          <label className="sr-only" htmlFor="name">Name</label>
          <input
            autoFocus
            className="border rounded-sm mr-2 p-3"
            name="name"
            onChange={(e) => setUsername(e.target.value)}
            onKeyUp={handleKeyPress}
            placeholder="email"
            type="text"
            value={email}
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
