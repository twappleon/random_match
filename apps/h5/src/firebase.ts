import { initializeApp } from 'firebase/app'
import { getAnalytics, isSupported } from 'firebase/analytics'
import {
  ConfirmationResult,
  GoogleAuthProvider,
  OAuthProvider,
  RecaptchaVerifier,
  User,
  createUserWithEmailAndPassword,
  getAuth,
  signInWithEmailAndPassword,
  signInWithPhoneNumber,
  signInWithPopup,
  signOut
} from 'firebase/auth'

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
export const firebaseAuthClient = firebaseApp ? getAuth(firebaseApp) : null

export async function initAnalytics() {
  if (firebaseApp && await isSupported()) {
    return getAnalytics(firebaseApp)
  }
  return null
}

function requireFirebaseAuth() {
  if (!firebaseAuthClient) {
    throw new Error('Firebase 尚未设置，请先配置 VITE_FIREBASE_* 环境变量')
  }
  return firebaseAuthClient
}

export async function loginWithEmailPassword(email: string, password: string): Promise<User> {
  const credential = await signInWithEmailAndPassword(requireFirebaseAuth(), email, password)
  return credential.user
}

export async function registerWithEmailPassword(email: string, password: string): Promise<User> {
  const credential = await createUserWithEmailAndPassword(requireFirebaseAuth(), email, password)
  return credential.user
}

export async function loginWithGoogle(): Promise<User> {
  const credential = await signInWithPopup(requireFirebaseAuth(), new GoogleAuthProvider())
  return credential.user
}

export async function loginWithApple(): Promise<User> {
  const provider = new OAuthProvider('apple.com')
  provider.addScope('email')
  provider.addScope('name')
  const credential = await signInWithPopup(requireFirebaseAuth(), provider)
  return credential.user
}

let phoneConfirmation: ConfirmationResult | null = null
let recaptchaVerifier: RecaptchaVerifier | null = null

export async function sendPhoneVerification(phoneNumber: string): Promise<void> {
  const auth = requireFirebaseAuth()
  if (!recaptchaVerifier) {
    recaptchaVerifier = new RecaptchaVerifier(auth, 'firebase-recaptcha', {
      size: 'invisible'
    })
  }
  phoneConfirmation = await signInWithPhoneNumber(auth, phoneNumber, recaptchaVerifier)
}

export async function confirmPhoneVerification(code: string): Promise<User> {
  if (!phoneConfirmation) {
    throw new Error('请先发送验证码')
  }
  const credential = await phoneConfirmation.confirm(code)
  phoneConfirmation = null
  return credential.user
}

export async function logoutFirebase(): Promise<void> {
  if (!firebaseAuthClient) return
  await signOut(firebaseAuthClient)
}
