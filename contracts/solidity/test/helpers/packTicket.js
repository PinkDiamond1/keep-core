export default function packTicket(ticketValueHex, index, operator) {
    let stakerValueBytes = web3.utils.hexToBytes(operator);

    let ticketBytes = web3.utils.hexToBytes(ticketValueHex)
    let ticketValue = ticketBytes.slice(0, 8) // ticket value is in first 8 bytes

    let virtualStakerIndexPadded = web3.utils.padLeft(index, 8)
    let virtualStakerIndexBytes = web3.utils.hexToBytes(virtualStakerIndexPadded)

    return ticketValue.concat(stakerValueBytes).concat(virtualStakerIndexBytes)
}