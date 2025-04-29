# Agentic Consensus  
## [A Discussion-Driven Layer 1 Blockchain for Autonomous Validation](https://gist.github.com/Deeptanshu-sankhwar/c432386deb0097f630134fffdb2abb47)

---

### Overview

**Agentic Consensus** is a prototype Layer 1 blockchain where **validator nodes are autonomous AI agents** capable of **discussing and voting** on incoming proposals.

Unlike traditional blockchains where validators simply verify transaction signatures or block hashes, Agentic Consensus validators **engage in structured discussions** to decide whether to approve or reject a proposed action. A proposal is committed to the blockchain if it achieves **at least 2/3rd validator approval**.

This system simulates decentralized decision-making for real-world scenarios that require reasoning and deliberation beyond simple transaction validation.

---

### Example Use Cases

Validator nodes in Agentic Consensus deliberate over various types of proposals, such as:

- **Research Paper Review:**  
  Agents read a submitted research paper and discuss its validity before approving its registration on-chain.

- **Political Discussions:**  
  Agents engage in structured debates regarding policy decisions or governance proposals.

- **Loan Application Evaluation:**  
  Agents analyze a loan applicant’s profile and collectively decide whether the loan should be approved.

In all cases, a proposal is registered only if **more than two-thirds (2/3)** of participating validator nodes vote to approve.

---

### How It Works

1. A **proposal** is submitted to the blockchain network.
2. All active **validator agents** receive the proposal.
3. Each validator:
   - Reviews the proposal
   - Engages in internal reasoning and discussion with peers (simulated)
   - Votes to either **approve** or **reject** the proposal
4. If **≥ 2/3** of validators approve, the proposal is committed as a transaction to the blockchain.
5. Otherwise, the proposal is rejected and not included in the chain.

---

### Architecture

```
+--------------------------+
|   Client (User/Submitter) |
+------------+-------------+
             |
      Submits Proposal
             |
+------------v-------------+
|      Validator Agents     |
|  - Receive Proposal       |
|  - Discuss and Reason     |
|  - Vote (Approve/Reject)  |
+------------+-------------+
             |
      Aggregate Votes
             |
+------------v-------------+
|       Blockchain Layer    |
|  - Commit if ≥2/3 approve  |
|  - Reject otherwise       |
+---------------------------+
```

---

### Quick Start

1. **Clone the Repository:**

   ```bash
   git clone https://github.com/Deeptanshu-sankhwar/agentic-consensus.git
   cd agentic-consensus
   ```

2. **Install Dependencies:**

   ```bash
   go mod tidy
   ```

3. **Run Genesis Validator Node Locally:**

   ```bash
   go run cmd/main.go
   ```

4. **Submit a Proposal via API:**  
   - POST `/api/transactions`
   - Validators will simulate discussion and vote.
   - Chain will automatically commit if the approval threshold is met.

---

### Future Work

- More realistic discussion models among validator agents
- Support for agent specialization based on proposal type (e.g., finance, governance, research)
- Integration with real-world document parsing for paper reviews and loan profile evaluation
- Robust peer-to-peer networking layer for validator communication
- Reputation systems to weigh votes differently based on past agent performance

---

### Project Status

> **Phase:**  
> Minimal Viable Prototype (MVP) — Validators reason about proposals and achieve 2/3rd voting consensus for block inclusion.

---