import unittest
import requests

BASE_URL = "http://localhost:8081"

class APITests(unittest.TestCase):

    @classmethod
    def setUpClass(cls):
        # Create a user "testuser" and get JWT token from /api/auth.
        auth_url = f"{BASE_URL}/api/auth"
        data = {"username": "testuser", "password": "testpass"}
        resp = requests.post(auth_url, json=data)
        assert resp.status_code == 200, f"Auth failed: {resp.text}"
        cls.token = resp.json().get("token")
        # Also create a recipient user for testing sendCoin
        data_recipient = {"username": "recipient", "password": "recipientpass"}
        resp_recipient = requests.post(auth_url, json=data_recipient)
        assert resp_recipient.status_code == 200, f"Recipient auth failed: {resp_recipient.text}"
        cls.recipient_username = "recipient"

    def auth_headers(self, token=None):
        if token is None:
            token = self.__class__.token
        return {"Authorization": f"Bearer {token}"}

    def test_auth_success(self):
        url = f"{BASE_URL}/api/auth"
        payload = {"username": "newuser", "password": "newpass"}
        resp = requests.post(url, json=payload)
        self.assertEqual(resp.status_code, 200)
        self.assertIn("token", resp.json())

    def test_auth_missing_fields(self):
        url = f"{BASE_URL}/api/auth"
        payload = {"username": "incompleteuser"}
        resp = requests.post(url, json=payload)
        self.assertEqual(resp.status_code, 400)

    def test_info_authorized(self):
        url = f"{BASE_URL}/api/info"
        headers = self.auth_headers()
        resp = requests.get(url, headers=headers)
        self.assertEqual(resp.status_code, 200)
        data = resp.json()
        self.assertIn("coins", data)
        self.assertIn("inventory", data)
        self.assertIn("coinHistory", data)

    def test_info_unauthorized(self):
        url = f"{BASE_URL}/api/info"
        resp = requests.get(url)
        self.assertEqual(resp.status_code, 401)

    def test_sendCoin_authorized(self):
        url = f"{BASE_URL}/api/sendCoin"
        headers = self.auth_headers()
        payload = {"toUser": self.__class__.recipient_username, "amount": 10}
        resp = requests.post(url, json=payload, headers=headers)
        self.assertEqual(resp.status_code, 200)

    def test_sendCoin_unauthorized(self):
        url = f"{BASE_URL}/api/sendCoin"
        payload = {"toUser": self.__class__.recipient_username, "amount": 10}
        resp = requests.post(url, json=payload)  # No Authorization header
        self.assertEqual(resp.status_code, 401)

    def test_buy_authorized(self):
        # Assumes that "testuser" has enough coins to buy "t-shirt"
        url = f"{BASE_URL}/api/buy/t-shirt"
        headers = self.auth_headers()
        resp = requests.get(url, headers=headers)
        self.assertEqual(resp.status_code, 200)

    def test_buy_unauthorized(self):
        url = f"{BASE_URL}/api/buy/t-shirt"
        resp = requests.get(url)
        self.assertEqual(resp.status_code, 401)

if __name__ == "__main__":
    unittest.main()
