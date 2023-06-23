import Image from 'next/image'

export default function Home() {
  return (
    <main>
      <section className='voltmetro'>
        <div>{"00.00"} V</div>
        <div>
          <h1>V</h1>
          <div></div>
        </div>
        <div>
          <div>.</div>
          <div>.</div>
        </div>
      </section>
    </main>
  )
}
