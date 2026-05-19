import { initializeApp } from 'firebase/app'
import { getAnalytics, isSupported } from 'firebase/analytics'

const firebaseConfig = {
  apiKey: import.meta.env.VITE_FIREBASE_API_KEY,
  authDomain: import.meta.env.VITE_FIREBASE_AUTH_DOMAIN,
  projectId: import.meta.env.VITE_FIREBASE_PROJECT_ID,
  storageBucket: import.meta.env.VITE_FIREBASE_STORAGE_BUCKET,
  messagingSenderId: import.meta.env.VITE_FIREBASE_MESSAGING_SENDER_ID,
  appId: import.meta.env.VITE_FIREBASE_APP_ID,
  measurementId: import.meta.env.VITE_FIREBASE_MEASUREMENT_ID
}

const hasFirebaseConfig = Boolean(firebaseConfig.apiKey && firebaseConfig.projectId && firebaseConfig.appId)

export const firebaseApp = hasFirebaseConfig ? initializeApp(firebaseConfig) : null

export async function initAnalytics() {
  if (firebaseApp && await isSupported()) {
    return getAnalytics(firebaseApp)
  }
  return null
}
