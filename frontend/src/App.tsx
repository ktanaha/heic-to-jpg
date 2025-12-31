import React from 'react'
import ImageConverter from './components/ImageConverter'
import './App.css'

function App() {
  return (
    <div className="App">
      <header className="App-header">
        <h1>HEIC to JPG Converter</h1>
        <p>HEICファイルをJPGに変換します</p>
      </header>
      <main>
        <ImageConverter />
      </main>
    </div>
  )
}

export default App
