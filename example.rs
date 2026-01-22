// Deriving a PDA to use as delegate
let (vault_pda, vault_bump) = Pubkey::find_program_address(
    &[b"vault", user.key().as_ref()],
    ctx.program_id
);

// The PDA can now act as a delegate
pub fn transfer_with_pda_delegate(ctx: Context<TransferWithPDA>, amount: u64) -> Result<()> {
    let seeds = &[
        b"vault",
        ctx.accounts.user.key().as_ref(),
        &[ctx.bumps.vault_pda],
    ];
    let signer = &[&seeds[..]];

    let cpi_accounts = Transfer {
        from: ctx.accounts.token_account.to_account_info(),
        to: ctx.accounts.destination.to_account_info(),
        authority: ctx.accounts.vault_pda.to_account_info(),
    };

    let cpi_program = ctx.accounts.token_program.to_account_info();
    let cpi_ctx = CpiContext::new_with_signer(cpi_program, cpi_accounts, signer);

    token::transfer(cpi_ctx, amount)?;
    Ok(())
}
