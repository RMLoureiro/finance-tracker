# frozen_string_literal: true

class AccountsController < ApplicationController
  protect_from_forgery with: :null_session
  before_action :require_login
  before_action :set_account, only: [:show]

  # GET /accounts
  def index
    accounts = current_user.accounts
    render json: accounts
  end

  # GET /accounts/:id
  def show
    render json: @account
  end

  # POST /accounts
  def create
    main = current_user.accounts.none?
    account = current_user.accounts.build(account_params)
    account.main = main
    if account.save
      render json: account, status: :created
    else
      render json: { errors: account.errors.full_messages }, status: :unprocessable_entity
    end
  end

  # GET /accounts/main
  def main
    account = current_user.accounts.find_by(main: true)
    if account
      transactions = account.transactions.order(created_at: :desc).limit(10)
      budgets = account.budgets
      render json: {
        account:             account,
        recent_transactions: transactions,
        budgets:             budgets
      }
    else
      render json: { error: "No main account found" }, status: :not_found
    end
  end

  private

  def set_account
    @account = current_user.accounts.find(params[:id])
  end

  def account_params
    params.expect(account: %i[name balance account_type])
  end
end
