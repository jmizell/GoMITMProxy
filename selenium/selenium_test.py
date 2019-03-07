from selenium import webdriver
from selenium.webdriver.common.desired_capabilities import DesiredCapabilities

import base64
import os

driver = webdriver.Remote(
    command_executor='http://127.0.0.1:4444/wd/hub',
    desired_capabilities=DesiredCapabilities.CHROME)

passing = False
try:
    try:
        os.remove("imageToSave.png")
    except Exception as e:
        print(e)

    driver.get("https://www.reddit.com")

    screenshot = driver.get_screenshot_as_base64()
    with open("/tmp/selenium_screenshot.png", "wb") as fh:
        fh.write(base64.b64decode(screenshot))

    print(driver.title)
    assert "the front page of the internet" in driver.title

    passing = True
except Exception as e:
    print(e.with_traceback())

driver.close()

if not passing:
    exit(1)
