"use client";
import React, { useState, useEffect } from 'react';

export default function Multimetro() {
  const [data, setData] = useState(0);
  const [socket, setSocket] = useState(null);

  const createSocket = () => {
    const newSocket = new WebSocket(`ws://192.168.0.172:3000/ws`);
    setSocket(newSocket);
    return () => {
      if (newSocket) {
        newSocket.close();
      }
    };
  };

  useEffect(() => {
    const cleanup = createSocket();
    return () => {
      cleanup();
    };
  }, []);

  useEffect(() => {
    const fetchData = (event) => {
      try {
        const responseData = JSON.parse(event.data);
        setData(Number(responseData));
      } catch (error) {
        console.error('Error al procesar el mensaje:', error);
      }
    };

    if (socket) {
      socket.onmessage = fetchData;
    }

    return () => {
      if (socket) {
        socket.onmessage = null;
      }
    };
  }, [socket]);


  return (
    <>
      <div>{(data/(255)*60).toFixed(3)} V</div>
      <div>
        <h1>V</h1>
        <div></div>
      </div>
      <div>
        <div><span>+</span></div>
        <div>-</div>
      </div>
    </>
  );
}
