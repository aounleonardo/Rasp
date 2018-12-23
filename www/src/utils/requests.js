export default async function raspRequest(endpoint, request, payload, callback) {
        fetch(endpoint+ request, payload)
            .then(res => res.json())
            .then(parsed => callback(parsed))
            .catch(err => console.log({err: err}))
}
