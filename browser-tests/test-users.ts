import * as fs from 'fs';
import * as path from 'path';
import { parse } from 'yaml';

interface Config {
  db: {
    admin: {
      username: string;
      email: string;
      password: string;
    };
  };
}

function loadConfig(): Config {
  const configPath = path.join(__dirname, '..', 'config.yaml');
  const configContent = fs.readFileSync(configPath, 'utf-8');
  return parse(configContent) as Config;
}

const config = loadConfig();

export const ADMIN_USER = {
  email: config.db.admin.email,
  password: config.db.admin.password,
};

const testRunId = Date.now();

export const TEST_USERS = {
  approved1: {
    username: `testuser1_${testRunId}`,
    email: `testuser1_${testRunId}@example.com`,
    password: 'TestPassword123!',
  },
  approved2: {
    username: `testuser2_${testRunId}`,
    email: `testuser2_${testRunId}@example.com`,
    password: 'TestPassword456!',
  },
  rejected: {
    username: `rejecteduser_${testRunId}`,
    email: `rejected_${testRunId}@example.com`,
    password: 'RejectedPass789!',
  },
};
