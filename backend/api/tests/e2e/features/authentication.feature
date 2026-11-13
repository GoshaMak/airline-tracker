Feature: User authentication and planned password rotation
  As a registered user
  I want a second factor delivered to my email
  So that a password alone cannot open a session

  Scenario: Login requires the email verification code
    Given a registered technical user
    When I submit the correct password
    Then the API asks for email verification
    When I submit the code from the verification email
    Then I receive an authenticated session

  Scenario: Repeated password failures lock the account and email recovery unlocks it
    Given a registered technical user
    When I submit an incorrect password five times
    Then the account is locked
    When I request account recovery
    And I submit the code from the recovery email
    Then the account is unlocked
    And the API asks for email verification
    When I submit the code from the verification email
    Then I receive an authenticated session

  Scenario: Repeated email verification failures lock the account
    Given a registered technical user
    When I submit the correct password
    And I submit an incorrect verification code five times
    Then the account is locked
    When I request account recovery
    And I submit the code from the recovery email
    Then the account is unlocked
    And the API asks for email verification
    When I submit the code from the verification email
    Then I receive an authenticated session

  Scenario: A signed-in user can rotate their password
    Given a registered technical user
    When I submit the correct password
    And I submit the code from the verification email
    Then I receive an authenticated session
    When I change my password
    Then the new password works and the old password is rejected
