/**
 * Lambda function invoked by AWS CloudWatch that forwards the logs to a syslog server.
 *
 * Forked from https://github.com/apiaryio/cloudwatch-to-papertrail
 * (license MIT, Copyright (c) 2016 Apiary Inc.), and updated so that it works reliably.
 */
const { promisify } = require("util");
const zlib = require("zlib");
const winston = require("winston");

// Register the winston papertrail transport
require("winston-papertrail").Papertrail;

/**
 * Checks that the given environment variable is present and returns its value.
 */
function mandatoryEnvVar(envVarName) {
  const value = process.env[envVarName];
  if (!value) {
    throw new Error(`environment variable ${envVarName} is mandatory`);
  }
  return value;
}

/** Syslog server host to send syslogs to. */
const SYSLOG_SERVER_HOST = mandatoryEnvVar("SYSLOG_SERVER_HOST");

/** Syslog server port to send syslogs to. */
const SYSLOG_SERVER_PORT = parseInt(mandatoryEnvVar("SYSLOG_SERVER_PORT"), 10);

/** Whether to disable sending the syslogs over TLS and use a raw TCP stream. */
const DISABLE_TLS = process.env["DISABLE_TLS"] === "1";

function getProgramFromLogStream(logStream) {
  const i = logStream.indexOf("/");
  if (i === -1) return logStream;
  return logStream.substr(0, i);
}

exports.handler = async function(event, context) {
  // By default, lambda wait for the event loop to empty before shutting down the function.
  // `true` is the default value, we put this here as a reminder that we make use of
  // this functionality.
  context.callbackWaitsForEmptyEventLoop = true;

  // We get the JSON log event data by un-gzipping the base64-encoded payload directly given
  // to the lambda function by CloudWatch
  const payload = new Buffer(event.awslogs.data, "base64");
  const result = await promisify(zlib.gunzip)(payload);
  const data = JSON.parse(result.toString("utf8"));

  // We open a papertrail Transport here. `winston-papertrail` is actually not specific
  // to Papertrail and can accommodate any syslog server.
  //
  // The Papertrail transport opens a TCP connection that is closed when all the logs are sent.
  // TCP connections could be reused inside the same lambda instance across invocations, but
  // this would mean changing the internals of `winston-papertrail` as then we wouldn't
  // flush the log messages properly to Papertrail without adapting the logic a bit.
  var transport = new winston.transports.Papertrail({
    host: SYSLOG_SERVER_HOST,
    port: SYSLOG_SERVER_PORT,
    disableTls: DISABLE_TLS,
    hostname: data.logGroup,
    program: getProgramFromLogStream(data.logStream),
    flushOnClose: true,
    logFormat: function(_level, message) {
      return message;
    }
  });
  var logger = new winston.Logger({ transports: [transport] });

  await new Promise((accept, reject) => {
    transport.on("error", err => {
      reject(err);
    });

    transport.on("connect", () => {
      for (const logEvent of data.logEvents) {
        logger.info(logEvent.message);
      }
      logger.close();
      accept();
    });
  });
};
