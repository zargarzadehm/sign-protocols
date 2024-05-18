import { Communicator } from '../../lib/abstract/Communicator';
import { EncryptionHandler } from '../../lib/abstract/EncryptionHandler';
import { DummyLogger } from '@rosen-bridge/abstract-logger';

export class TestCommunicator extends Communicator {
  constructor(
    signer: EncryptionHandler,
    submitMessage: (msg: string, peers: Array<string>) => unknown,
    guardPks: Array<string>,
  ) {
    super(new DummyLogger(), signer, submitMessage, guardPks);
  }

  testSendMessage = (
    messageType: string,
    payload: any,
    peers: Array<string>,
  ) => {
    return this.sendMessage(messageType, payload, peers);
  };

  processMessage = jest.fn();

  mockedGetDate = () => this.getDate();
}
