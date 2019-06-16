const request = require("request-promise");
const log4js = require('log4js');

const SERVER_URL = "http://68.183.40.252:3000";

const logger = log4js.getLogger();
logger.level = 'debug';

const delay = ms => new Promise(res => setTimeout(res, ms));

async function main() {
  try {
    const pendingTxs = {};
    const sessionId = String(Math.floor(Math.random() * 300 + 1));
    logger.info(`Start test client (session is ${sessionId}), getting current contract state...`);

    let res = await request.get(`${SERVER_URL}/state/latest`);
    res = JSON.parse(res);

    logger.info(`Current contract state is ${res.value}`);

    const countFrom1To3 = Math.floor(Math.random() * 3 + 1);
    const operation = (Math.random() >= 0.5) ? "increment" : "decrement";
    logger.info(`Sending ${countFrom1To3} operation of type ${operation}`);

    for (let i = 0; i < countFrom1To3; i++) {
      let res = await request.get(`${SERVER_URL}/state/${operation}?session=${sessionId}`);
      res = JSON.parse(res);
      pendingTxs[res.hash] = true;
    }

    logger.info("Start watch for txs states");

    while (Object.keys(pendingTxs).length !== 0) {
      logger.debug("Updating tx statuses...");

      let watchedTxs = await request.get(`${SERVER_URL}/tx/session?session=${sessionId}`);
      watchedTxs = JSON.parse(watchedTxs);

      for (let i = 0; i < watchedTxs.length; i++) {
        if (watchedTxs[i].pending || !pendingTxs[watchedTxs[i].hash]) {
          continue
        }

        // warn level only for visibility
        logger.warn(`Tx ${watchedTxs[i].hash} done...`);
        delete pendingTxs[watchedTxs[i].hash];
      }

      // sleep for 5s
      await delay(5000);
    }

    logger.info("All req successfully finished");
  } catch (err) {
    logger.error(err);
  }
}

// will start with timeout in range from 0s to 1s 
setTimeout(main, Math.random() * 1000);