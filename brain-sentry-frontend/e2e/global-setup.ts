import { ApiHelper } from "./helpers/api.helper";

async function globalSetup() {
  const api = new ApiHelper();
  try {
    await api.ensureDemoUser();
    console.log("Global setup: demo user ensured");
  } catch (err) {
    console.warn("Global setup warning:", err);
  }
}

export default globalSetup;
