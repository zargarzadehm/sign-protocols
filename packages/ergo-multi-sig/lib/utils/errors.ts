export class CommitmentMisMatch extends Error {
  constructor(msg: string) {
    super('CommitmentMismatch: ' + msg);
  }
}
