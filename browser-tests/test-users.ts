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

export const TEST_USERS = {
  approved1: {
    username: 'testuser1',
    email: 'testuser1@example.com',
    password: 'TestPassword123!',
  },
  approved2: {
    username: 'testuser2',
    email: 'testuser2@example.com',
    password: 'TestPassword456!',
  },
};
