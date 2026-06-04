// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract TrustChain {
    struct TxRecord {
        string txHash;
        string fromAddr;
        string toAddr;
        uint256 value;
        uint256 fee;
        uint256 gasUsed;
        uint256 timestamp;
        uint256 blockNumber;
        bool isFraud;
        string verdict;
        string flagReason;
        uint256 riskScore;
        string data;
        string status;
    }

    struct CorrectionRecord {
        string txHash;
        bool isCorrected;
        string actualStatus;
        string reason;
        string correctedBy;
        uint256 updatedAt;
    }

    mapping(string => TxRecord) private records;
    string[] private allHashes;
    uint256 public fraudCount;

    // Mapping dari block number ke array txHash di block tersebut
    mapping(uint256 => string[]) private blockToTxHashes;
    mapping(string => CorrectionRecord) private corrections;

    event TxRecordCreated(string txHash, string fromAddr, string toAddr, uint256 value, uint256 blockNumber);
    event TxRecordUpdated(string txHash, bool isFraud, uint256 riskScore, string verdict, string flagReason, string status);
    event TxCorrectionUpdated(string txHash, bool isCorrected, string actualStatus, string reason, string correctedBy, uint256 updatedAt);

    function recordTransaction(
        string memory _txHash,
        string memory _from,
        string memory _to,
        uint256 _value,
        uint256 _fee,
        uint256 _gasUsed,
        string memory _data
    ) public {
        records[_txHash] = TxRecord({
            txHash: _txHash,
            fromAddr: _from,
            toAddr: _to,
            value: _value,
            fee: _fee,
            gasUsed: _gasUsed,
            timestamp: block.timestamp,
            blockNumber: block.number,
            isFraud: false,
            verdict: "pending",
            flagReason: "",
            riskScore: 0,
            data: _data,
            status: "pending"
        });
        allHashes.push(_txHash);
        blockToTxHashes[block.number].push(_txHash);
        emit TxRecordCreated(_txHash, _from, _to, _value, block.number);
    }

    function updateTransactionPrediction(
        string memory _txHash,
        bool _isFraud,
        uint256 _riskScore,
        string memory _verdict,
        string memory _flagReason
    ) public {
        TxRecord storage r = records[_txHash];
        r.isFraud = _isFraud;
        r.riskScore = _riskScore;
        r.verdict = _verdict;
        r.flagReason = _flagReason;
        r.status = "success";
        
        if (_isFraud) {
            fraudCount++;
        }
        
        emit TxRecordUpdated(_txHash, _isFraud, _riskScore, _verdict, _flagReason, "success");
    }

    function addCorrection(
        string memory _txHash,
        bool _isCorrected,
        string memory _actualStatus,
        string memory _reason,
        string memory _correctedBy
    ) public {
        corrections[_txHash] = CorrectionRecord({
            txHash: _txHash,
            isCorrected: _isCorrected,
            actualStatus: _actualStatus,
            reason: _reason,
            correctedBy: _correctedBy,
            updatedAt: block.timestamp
        });
        emit TxCorrectionUpdated(_txHash, _isCorrected, _actualStatus, _reason, _correctedBy, block.timestamp);
    }

    function getTransactionCorrection(string memory _txHash) public view returns (
        bool isCorrected,
        string memory actualStatus,
        string memory reason,
        string memory correctedBy,
        uint256 updatedAt
    ) {
        CorrectionRecord memory c = corrections[_txHash];
        return (
            c.isCorrected, c.actualStatus, c.reason, c.correctedBy, c.updatedAt
        );
    }

    function getTransactionBase(string memory _txHash) public view returns (
        string memory txHash,
        string memory fromAddr,
        string memory toAddr,
        uint256 value,
        uint256 fee,
        uint256 gasUsed,
        uint256 timestamp,
        uint256 blockNumber
    ) {
        TxRecord memory r = records[_txHash];
        require(bytes(r.txHash).length > 0, "Transaction not found");
        return (
            r.txHash, r.fromAddr, r.toAddr, r.value, r.fee, r.gasUsed, r.timestamp, r.blockNumber
        );
    }

    function getTransactionPrediction(string memory _txHash) public view returns (
        bool isFraud,
        string memory verdict,
        string memory flagReason,
        uint256 riskScore,
        string memory data,
        string memory status
    ) {
        TxRecord memory r = records[_txHash];
        require(bytes(r.txHash).length > 0, "Transaction not found");
        return (
            r.isFraud, r.verdict, r.flagReason, r.riskScore, r.data, r.status
        );
    }

    function getTransactionCount() public view returns (uint256) {
        return allHashes.length;
    }

    function getStats() public view returns (uint256 totalTxs, uint256 totalFraud) {
        return (allHashes.length, fraudCount);
    }

    function getTransactionHashAtIndex(uint256 index) public view returns (string memory) {
        require(index < allHashes.length, "Index out of bounds");
        return allHashes[index];
    }

    function getBlockTransactionCount(uint256 _blockNumber) public view returns (uint256) {
        return blockToTxHashes[_blockNumber].length;
    }

    function getBlockTransactionHashAtIndex(uint256 _blockNumber, uint256 _index) public view returns (string memory) {
        require(_index < blockToTxHashes[_blockNumber].length, "Index out of bounds");
        return blockToTxHashes[_blockNumber][_index];
    }
}

