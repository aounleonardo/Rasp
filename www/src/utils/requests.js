export async function raspRequest(endpoint, request, payload, callback) {
    const shipment = (payload) ?
        {
            method: 'post',
            body: JSON.stringify(payload),
        } :
        null;
    fetch(endpoint + request, shipment)
        .then(res => res.json())
        .then(parsed => callback(parsed))
        .catch(err => console.log({err: err}))
}
