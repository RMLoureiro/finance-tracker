require 'rails_helper'

RSpec.describe User, type: :model do
  describe "structure and validations" do
    it "has all columns" do
      expect(subject).to have_db_column(:username)
      expect(subject).to have_db_column(:name)
      expect(subject).to have_db_column(:password_digest)
    end
  end
end
