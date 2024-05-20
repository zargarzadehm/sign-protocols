import { EcdsaSigner, ECDSA, GuardDetection } from '../../lib';

describe('EcdsaSigner', () => {
  const currentTime = 1686286005068;

  describe('signPromised', () => {
    /**
     * @target TssSigner.signPromised should throw error when derivationPath is not defined
     * @dependencies
     * @scenario
     * - generate EcdsaSigner object using mocked data
     * - call signPromised with undefined derivationPath and check thrown exception
     * @expected
     * - it should throw Error
     */
    it('should throw error when derivationPath is not defined', async () => {
      const sk = await ECDSA.randomKey();
      const signer = new ECDSA(sk);
      jest.restoreAllMocks();
      jest.spyOn(Date, 'now').mockReturnValue(currentTime);
      const detection = new GuardDetection({
        signer: signer,
        guardsPublicKey: [],
        submit: jest.fn(),
        getPeerId: () => Promise.resolve('myPeerId'),
      });
      const ecdsaSigner = new EcdsaSigner({
        submitMsg: jest.fn(),
        callbackUrl: '',
        secret: sk,
        detection: detection,
        guardsPk: [],
        tssApiUrl: '',
        getPeerId: () => Promise.resolve('myPeerId'),
        shares: [],
      });

      await expect(async () => {
        await ecdsaSigner.signPromised('message', 'chainCode', undefined);
      }).rejects.toThrow(Error);
    });
  });
});
