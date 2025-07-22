# frozen_string_literal: true

require "rails_helper"

RSpec.describe User do
  describe "structure and validations" do
    it "has all columns" do
      expect(subject).to have_db_column(:username)
      expect(subject).to have_db_column(:name)
      expect(subject).to have_db_column(:password_digest)
    end

    it "validates presence of username, name, and password" do
      user = described_class.new
      expect(user).not_to be_valid
      expect(user.errors[:username]).to include("can't be blank")
      expect(user.errors[:name]).to include("can't be blank")
      expect(user.errors[:password]).to include("can't be blank")
    end

    it "validates uniqueness of username" do
      described_class.create!(username: "uniqueuser", name: "Name", password: "Password123!")
      user2 = described_class.new(username: "uniqueuser", name: "Other", password: "Password123!")
      expect(user2).not_to be_valid
      expect(user2.errors[:username]).to include("has already been taken")
    end

    it "creates a user with valid params" do
      user = described_class.new(username: "newuser", name: "New User", password: "Password123!")
      expect(user).to be_valid
    end
  end
end
