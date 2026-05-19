import { createReadStream, existsSync, statSync } from 'node:fs'
import { extname, join, normalize, resolve } from 'node:path'
import { createServer } from 'node:http'

const root = resolve(process.cwd(), 'apps/h5')
const port = Number(process.env.PORT ?? 5173)

const types = {
  '.css': 'text/css; charset=utf-8',
  '.html': 'text/html; charset=utf-8',
  '.js': 'text/javascript; charset=utf-8',
  '.json': 'application/json; charset=utf-8',
  '.ts': 'text/plain; charset=utf-8',
  '.vue': 'text/plain; charset=utf-8'
}

createServer((req, res) => {
  const url = new URL(req.url ?? '/', `http://${req.headers.host}`)
  const requested = url.pathname === '/' ? '/preview.html' : url.pathname
  const file = normalize(join(root, requested))

  if (!file.startsWith(root) || !existsSync(file) || !statSync(file).isFile()) {
    res.writeHead(404, { 'content-type': 'text/plain; charset=utf-8' })
    res.end('Not found')
    return
  }

  res.writeHead(200, { 'content-type': types[extname(file)] ?? 'application/octet-stream' })
  createReadStream(file).pipe(res)
}).listen(port, '0.0.0.0', () => {
  console.log(`Preview: http://localhost:${port}`)
})
