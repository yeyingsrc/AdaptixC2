#include <Agent/Agent.h>
#include <Agent/AgentTableWidgetItem.h>
#include <UI/Widgets/SessionsTableWidget.h>
#include <UI/Widgets/BrowserFilesWidget.h>
#include <UI/Widgets/BrowserProcessWidget.h>
#include <UI/Widgets/ConsoleWidget.h>
#include <UI/Widgets/TerminalWidget.h>
#include <UI/Widgets/AdaptixWidget.h>
#include <UI/Widgets/TasksWidget.h>
#include <UI/Dialogs/DialogTunnel.h>
#include <Client/AxScript/AxScriptManager.h>
#include <Client/Requestor.h>
#include <Client/Settings.h>
#include <Client/TunnelEndpoint.h>
#include <Client/AuthProfile.h>
#include <MainAdaptix.h>

SessionsTableWidget::SessionsTableWidget( AdaptixWidget* w )
{
    this->adaptixWidget = w;

    this->createUI();

    connect( tableWidget, &QTableWidget::doubleClicked,              this, &SessionsTableWidget::handleTableDoubleClicked );
    connect( tableWidget, &QTableWidget::customContextMenuRequested, this, &SessionsTableWidget::handleSessionsTableMenu );
    connect( tableWidget, &QTableWidget::itemSelectionChanged,       this, [this](){tableWidget->setFocus();} );

#if QT_VERSION >= QT_VERSION_CHECK(6, 7, 0)
    connect( checkOnlyActive, &QCheckBox::checkStateChanged, this, &SessionsTableWidget::onFilterUpdate);
#else
    connect( checkOnlyActive, &QCheckBox::stateChanged, this, &SessionsTableWidget::onFilterUpdate);
#endif

    connect( inputFilter1, &QLineEdit::textChanged,  this, &SessionsTableWidget::onFilterUpdate);
    connect( inputFilter2, &QLineEdit::textChanged,  this, &SessionsTableWidget::onFilterUpdate);
    connect( inputFilter3, &QLineEdit::textChanged,  this, &SessionsTableWidget::onFilterUpdate);
    connect( hideButton,   &ClickableLabel::clicked, this, &SessionsTableWidget::toggleSearchPanel);

    shortcutSearch = new QShortcut(QKeySequence("Ctrl+F"), tableWidget);
    shortcutSearch->setContext(Qt::WidgetShortcut);
    connect(shortcutSearch, &QShortcut::activated, this, &SessionsTableWidget::toggleSearchPanel);


}

SessionsTableWidget::~SessionsTableWidget() = default;

void SessionsTableWidget::createUI()
{
    auto horizontalSpacer1 = new QSpacerItem(40, 20, QSizePolicy::Expanding, QSizePolicy::Minimum);
    auto horizontalSpacer2 = new QSpacerItem(40, 20, QSizePolicy::Expanding, QSizePolicy::Minimum);

    searchWidget = new QWidget(this);
    searchWidget->setVisible(false);

    checkOnlyActive = new QCheckBox("Only active");

    inputFilter1 = new QLineEdit(searchWidget);
    inputFilter1->setPlaceholderText("filter1");
    inputFilter1->setMaximumWidth(200);

    inputFilter2 = new QLineEdit(searchWidget);
    inputFilter2->setPlaceholderText("or filter2");
    inputFilter2->setMaximumWidth(200);

    inputFilter3 = new QLineEdit(searchWidget);
    inputFilter3->setPlaceholderText("or filter3");
    inputFilter3->setMaximumWidth(200);

    hideButton = new ClickableLabel("X");
    hideButton->setCursor( Qt::PointingHandCursor );

    searchLayout = new QHBoxLayout(searchWidget);
    searchLayout->setContentsMargins(0, 0, 0, 0);
    searchLayout->setSpacing(4);
    searchLayout->addSpacerItem(horizontalSpacer1);
    searchLayout->addWidget(checkOnlyActive);
    searchLayout->addWidget(inputFilter1);
    searchLayout->addWidget(inputFilter2);
    searchLayout->addWidget(inputFilter3);
    searchLayout->addWidget(hideButton);
    searchLayout->addSpacerItem(horizontalSpacer2);

    tableWidget = new QTableWidget( this );
    tableWidget->setColumnCount( ColumnCount );
    tableWidget->setContextMenuPolicy( Qt::CustomContextMenu );
    tableWidget->setAutoFillBackground( false );
    tableWidget->setShowGrid( false );
    tableWidget->setSortingEnabled( true );
    tableWidget->setWordWrap( false );
    tableWidget->setCornerButtonEnabled( false );
    tableWidget->setSelectionBehavior( QAbstractItemView::SelectRows );
    tableWidget->setFocusPolicy( Qt::NoFocus );
    tableWidget->setAlternatingRowColors( true );
    tableWidget->horizontalHeader()->setSectionResizeMode( QHeaderView::Stretch );
    tableWidget->horizontalHeader()->setCascadingSectionResizes( true );
    tableWidget->horizontalHeader()->setHighlightSections( false );
    tableWidget->verticalHeader()->setVisible( false );

    tableWidget->setHorizontalHeaderItem( ColumnAgentID,   new QTableWidgetItem( "ID" ) );
    tableWidget->setHorizontalHeaderItem( ColumnAgentType, new QTableWidgetItem( "Type" ) );
    tableWidget->setHorizontalHeaderItem( ColumnListener,  new QTableWidgetItem( "Listener" ) );
    tableWidget->setHorizontalHeaderItem( ColumnExternal,  new QTableWidgetItem( "External" ) );
    tableWidget->setHorizontalHeaderItem( ColumnInternal,  new QTableWidgetItem( "Internal" ) );
    tableWidget->setHorizontalHeaderItem( ColumnDomain,    new QTableWidgetItem( "Domain" ) );
    tableWidget->setHorizontalHeaderItem( ColumnComputer,  new QTableWidgetItem( "Computer" ) );
    tableWidget->setHorizontalHeaderItem( ColumnUser,      new QTableWidgetItem( "User" ) );
    tableWidget->setHorizontalHeaderItem( ColumnOs,        new QTableWidgetItem( "OS" ) );
    tableWidget->setHorizontalHeaderItem( ColumnProcess,   new QTableWidgetItem( "Process" ) );
    tableWidget->setHorizontalHeaderItem( ColumnProcessId, new QTableWidgetItem( "PID" ) );
    tableWidget->setHorizontalHeaderItem( ColumnThreadId,  new QTableWidgetItem( "TID" ) );
    tableWidget->setHorizontalHeaderItem( ColumnTags,      new QTableWidgetItem( "Tags" ) );
    tableWidget->setHorizontalHeaderItem( ColumnLast,      new QTableWidgetItem( "Last" ) );
    tableWidget->setHorizontalHeaderItem( ColumnSleep,     new QTableWidgetItem( "Sleep" ) );

    tableWidget->setItemDelegate(new PaddingDelegate(tableWidget));
    this->UpdateColumnsVisible();

    mainGridLayout = new QGridLayout( this );
    mainGridLayout->setContentsMargins( 0, 0,  0, 0);
    mainGridLayout->addWidget( searchWidget, 0, 0, 1, 1);
    mainGridLayout->addWidget( tableWidget,  1, 0, 1, 1);
}

bool SessionsTableWidget::filterItem(const AgentData &agent) const
{
    if ( !this->searchWidget->isVisible() )
        return true;

    if (this->checkOnlyActive->isChecked()) {
        if ( agent.Mark == "Terminated" || agent.Mark == "Inactive" || agent.Mark == "Disconnect" )
            return false;
    }

    QString username = agent.Username;
    if ( agent.Elevated )
        username = "* " + username;

    bool found = true;

    QString filter1 = this->inputFilter1->text();
    if( !filter1.isEmpty() ) {
        if ( agent.Id.contains(filter1, Qt::CaseInsensitive) ||
             agent.Name.contains(filter1, Qt::CaseInsensitive) ||
             agent.Listener.contains(filter1, Qt::CaseInsensitive) ||
             agent.ExternalIP.contains(filter1, Qt::CaseInsensitive) ||
             agent.InternalIP.contains(filter1, Qt::CaseInsensitive) ||
             agent.Process.contains(filter1, Qt::CaseInsensitive) ||
             agent.OsDesc.contains(filter1, Qt::CaseInsensitive) ||
             agent.Domain.contains(filter1, Qt::CaseInsensitive) ||
             agent.Computer.contains(filter1, Qt::CaseInsensitive) ||
             username.contains(filter1, Qt::CaseInsensitive) ||
             agent.Tags.contains(filter1, Qt::CaseInsensitive)
        )
            return true;
        else
            found = false;
    }

    QString filter2 = this->inputFilter2->text();
    if( !filter2.isEmpty() ) {
        if ( agent.Id.contains(filter2, Qt::CaseInsensitive) ||
             agent.Name.contains(filter2, Qt::CaseInsensitive) ||
             agent.Listener.contains(filter2, Qt::CaseInsensitive) ||
             agent.ExternalIP.contains(filter2, Qt::CaseInsensitive) ||
             agent.InternalIP.contains(filter2, Qt::CaseInsensitive) ||
             agent.Process.contains(filter2, Qt::CaseInsensitive) ||
             agent.OsDesc.contains(filter2, Qt::CaseInsensitive) ||
             agent.Domain.contains(filter2, Qt::CaseInsensitive) ||
             agent.Computer.contains(filter2, Qt::CaseInsensitive) ||
             username.contains(filter2, Qt::CaseInsensitive) ||
             agent.Tags.contains(filter2, Qt::CaseInsensitive)
        )
            return true;
        else
            found = false;
    }

    QString filter3 = this->inputFilter3->text();
    if( !filter3.isEmpty() ) {
        if ( agent.Id.contains(filter3, Qt::CaseInsensitive) ||
             agent.Name.contains(filter3, Qt::CaseInsensitive) ||
             agent.Listener.contains(filter3, Qt::CaseInsensitive) ||
             agent.ExternalIP.contains(filter3, Qt::CaseInsensitive) ||
             agent.InternalIP.contains(filter3, Qt::CaseInsensitive) ||
             agent.Process.contains(filter3, Qt::CaseInsensitive) ||
             agent.OsDesc.contains(filter3, Qt::CaseInsensitive) ||
             agent.Domain.contains(filter3, Qt::CaseInsensitive) ||
             agent.Computer.contains(filter3, Qt::CaseInsensitive) ||
             username.contains(filter3, Qt::CaseInsensitive) ||
             agent.Tags.contains(filter3, Qt::CaseInsensitive)
        )
            return true;
        else
            found = false;
    }

    return found;
}

void SessionsTableWidget::addTableItem(const Agent* newAgent) const
{
    if( tableWidget->rowCount() < 1 )
        tableWidget->setRowCount( 1 );
    else
        tableWidget->setRowCount( tableWidget->rowCount() + 1 );

    bool isSortingEnabled = tableWidget->isSortingEnabled();
    tableWidget->setSortingEnabled( false );
    tableWidget->setItem( tableWidget->rowCount() - 1, ColumnAgentID,   newAgent->item_Id );
    tableWidget->setItem( tableWidget->rowCount() - 1, ColumnAgentType, newAgent->item_Type );
    tableWidget->setItem( tableWidget->rowCount() - 1, ColumnListener,  newAgent->item_Listener );
    tableWidget->setItem( tableWidget->rowCount() - 1, ColumnExternal,  newAgent->item_External );
    tableWidget->setItem( tableWidget->rowCount() - 1, ColumnInternal,  newAgent->item_Internal );
    tableWidget->setItem( tableWidget->rowCount() - 1, ColumnDomain,    newAgent->item_Domain );
    tableWidget->setItem( tableWidget->rowCount() - 1, ColumnComputer,  newAgent->item_Computer );
    tableWidget->setItem( tableWidget->rowCount() - 1, ColumnUser,      newAgent->item_Username );
    tableWidget->setItem( tableWidget->rowCount() - 1, ColumnOs,        newAgent->item_Os );
    tableWidget->setItem( tableWidget->rowCount() - 1, ColumnProcess,   newAgent->item_Process );
    tableWidget->setItem( tableWidget->rowCount() - 1, ColumnProcessId, newAgent->item_Pid );
    tableWidget->setItem( tableWidget->rowCount() - 1, ColumnThreadId,  newAgent->item_Tid );
    tableWidget->setItem( tableWidget->rowCount() - 1, ColumnTags,      newAgent->item_Tags );
    tableWidget->setItem( tableWidget->rowCount() - 1, ColumnLast,      newAgent->item_Last );
    tableWidget->setItem( tableWidget->rowCount() - 1, ColumnSleep,     newAgent->item_Sleep );
    tableWidget->setSortingEnabled( isSortingEnabled );

    this->UpdateColumnsWidth();
}



void SessionsTableWidget::AddAgentItem( Agent* newAgent ) const
{
    if ( adaptixWidget->AgentsMap.contains(newAgent->data.Id) )
        return;

    adaptixWidget->AgentsMap[ newAgent->data.Id ] = newAgent;
    adaptixWidget->AgentsVector.push_back(newAgent->data.Id);

    if( !this->filterItem(newAgent->data) )
        return;

    this->addTableItem(newAgent);
}

void SessionsTableWidget::RemoveAgentItem(const QString &agentId) const
{
    if (!adaptixWidget->AgentsMap.contains(agentId))
        return;

    Agent* agent = adaptixWidget->AgentsMap[agentId];
    adaptixWidget->AgentsMap.remove(agentId);
    adaptixWidget->AgentsVector.removeOne(agentId);

    if (agent->Console)
        delete agent->Console;
    if (agent->FileBrowser)
        delete agent->FileBrowser;
    if (agent->ProcessBrowser)
        delete agent->ProcessBrowser;
    if (agent->Terminal)
        delete agent->Terminal;
    delete agent;

    for( int rowIndex = 0 ; rowIndex < tableWidget->rowCount() ; rowIndex++ ) {
        if (agentId == tableWidget->item( rowIndex, ColumnAgentID )->text()) {
            tableWidget->removeRow(rowIndex);
            break;
        }
    }
}

void SessionsTableWidget::SetData() const
{
    this->ClearTableContent();

    for (int i = 0; i < adaptixWidget->AgentsVector.size(); i++ ) {
        QString agentId = adaptixWidget->AgentsVector[i];
        Agent* agent = adaptixWidget->AgentsMap[agentId];
        if ( agent && agent->show && this->filterItem(agent->data) )
            this->addTableItem(agent);
    }
}

void SessionsTableWidget::UpdateColumnsVisible() const
{
    for (int i = 0; i < 15; i++) {
        if (GlobalClient->settings->data.SessionsTableColumns[i])
            tableWidget->showColumn(i);
        else
            tableWidget->hideColumn(i);
    }
}

void SessionsTableWidget::UpdateColumnsWidth() const
{
    tableWidget->horizontalHeader()->setSectionResizeMode(QHeaderView::ResizeToContents);
    tableWidget->horizontalHeader()->setSectionResizeMode(ColumnTags, QHeaderView::Stretch);

    int wDomain   = tableWidget->columnWidth(ColumnDomain);
    int wComputer = tableWidget->columnWidth(ColumnDomain);
    int wUser     = tableWidget->columnWidth(ColumnDomain);

    tableWidget->horizontalHeader()->setSectionResizeMode(ColumnDomain,   QHeaderView::Interactive);
    tableWidget->horizontalHeader()->setSectionResizeMode(ColumnComputer, QHeaderView::Interactive);
    tableWidget->horizontalHeader()->setSectionResizeMode(ColumnUser,     QHeaderView::Interactive);

    tableWidget->setColumnWidth(ColumnDomain, wDomain);
    tableWidget->setColumnWidth(ColumnDomain, wComputer);
    tableWidget->setColumnWidth(ColumnDomain, wUser);
}

void SessionsTableWidget::ClearTableContent() const
{
    for (int row = tableWidget->rowCount() - 1; row >= 0; row--) {
        for (int col = 0; col < tableWidget->columnCount(); ++col)
            tableWidget->takeItem(row, col);

        tableWidget->removeRow(row);
    }
}

void SessionsTableWidget::Clear() const
{
    adaptixWidget->AgentsVector.clear();

    for (auto agentId : adaptixWidget->AgentsMap.keys()) {
        Agent* agent = adaptixWidget->AgentsMap[agentId];
        adaptixWidget->AgentsMap.remove(agentId);
        delete agent->Console;
        delete agent->FileBrowser;
        delete agent->ProcessBrowser;
        delete agent->Terminal;
        delete agent;
    }

    this->ClearTableContent();

    checkOnlyActive->setChecked(false);
    inputFilter1->clear();
    inputFilter2->clear();
    inputFilter3->clear();
}

/// SLOTS

void SessionsTableWidget::toggleSearchPanel() const
{
    if (this->searchWidget->isVisible())
        this->searchWidget->setVisible(false);
    else
        this->searchWidget->setVisible(true);

    this->SetData();
}

void SessionsTableWidget::handleTableDoubleClicked(const QModelIndex &index) const
{
    QString AgentId = tableWidget->item(index.row(),0)->text();

    adaptixWidget->LoadConsoleUI(AgentId);
}

void SessionsTableWidget::onFilterUpdate() const { this->SetData(); }

/// Menu

void SessionsTableWidget::handleSessionsTableMenu(const QPoint &pos)
{
    if ( !tableWidget->itemAt(pos) )
        return;

    QStringList agentIds;
    for( int rowIndex = 0 ; rowIndex < tableWidget->rowCount() ; rowIndex++ ) {
        if ( tableWidget->item(rowIndex, 0)->isSelected() ) {
            QString agentId = tableWidget->item( rowIndex, ColumnAgentID )->text();
            agentIds.append(agentId);
        }
    }



    auto agentMenu = QMenu("Agent");
    agentMenu.addAction("Execute command", this, &SessionsTableWidget::actionExecuteCommand);
    agentMenu.addAction("Task manager", this, &SessionsTableWidget::actionTasksBrowserOpen);
    agentMenu.addSeparator();

    int agentCount = adaptixWidget->ScriptManager->AddMenuSession(&agentMenu, "SessionAgent", agentIds);
    if (agentCount > 0)
        agentMenu.addSeparator();

    agentMenu.addAction("Remove console data", this, &SessionsTableWidget::actionConsoleDelete);
    agentMenu.addAction("Remove from server", this, &SessionsTableWidget::actionAgentRemove);



    auto sessionMenu = QMenu("Session");
    sessionMenu.addAction("Mark as Active",   this, &SessionsTableWidget::actionMarkActive);
    sessionMenu.addAction("Mark as Inactive", this, &SessionsTableWidget::actionMarkInactive);
    sessionMenu.addSeparator();
    sessionMenu.addAction("Set items color", this, &SessionsTableWidget::actionItemColor);
    sessionMenu.addAction("Set text color",  this, &SessionsTableWidget::actionTextColor);
    sessionMenu.addAction("Reset color",     this, &SessionsTableWidget::actionColorReset);
    sessionMenu.addSeparator();
    sessionMenu.addAction( "Hide on client", this, &SessionsTableWidget::actionItemHide);



    auto ctxMenu = QMenu();
    ctxMenu.addAction("Console", this, &SessionsTableWidget::actionConsoleOpen);
    ctxMenu.addSeparator();
    ctxMenu.addMenu(&agentMenu);

    auto browserMenu = QMenu("Browsers");
    int browserCount = adaptixWidget->ScriptManager->AddMenuSession(&browserMenu, "SessionBrowser", agentIds);
    if (browserCount > 0)
        ctxMenu.addMenu(&browserMenu);

    auto accessMenu = QMenu("Access");
    int accessCount = adaptixWidget->ScriptManager->AddMenuSession(&accessMenu, "SessionAccess", agentIds);
    if (accessCount > 0)
        ctxMenu.addMenu(&accessMenu);

    adaptixWidget->ScriptManager->AddMenuSession(&ctxMenu, "SessionMain", agentIds);

    ctxMenu.addSeparator();
    ctxMenu.addMenu(&sessionMenu);
    ctxMenu.addAction("Set tag", this, &SessionsTableWidget::actionItemTag);
    ctxMenu.addAction("Show all items", this, &SessionsTableWidget::actionItemsShowAll);

    ctxMenu.exec(tableWidget->horizontalHeader()->viewport()->mapToGlobal(pos));
}

void SessionsTableWidget::actionConsoleOpen() const
{
    for( int rowIndex = 0 ; rowIndex < tableWidget->rowCount() ; rowIndex++ ) {
        if ( tableWidget->item(rowIndex, 0)->isSelected() ) {
            auto agentId = tableWidget->item( rowIndex, ColumnAgentID )->text();
            adaptixWidget->LoadConsoleUI(agentId);
        }
    }
}

void SessionsTableWidget::actionExecuteCommand()
{
    QStringList listId;
    for( int rowIndex = 0 ; rowIndex < tableWidget->rowCount() ; rowIndex++ ) {
        if ( tableWidget->item(rowIndex, 0)->isSelected() ) {
            auto agentId = tableWidget->item( rowIndex, ColumnAgentID )->text();
            listId.append(agentId);
        }
    }

    if(listId.empty())
        return;

    bool ok = false;
    QString cmd = QInputDialog::getText(this,"Execute Command", "Command", QLineEdit::Normal, "", &ok);
    if (!ok)
        return;

    for(auto id : listId) {
        adaptixWidget->AgentsMap[id]->Console->SetInput(cmd);
        adaptixWidget->AgentsMap[id]->Console->processInput();
    }
}

void SessionsTableWidget::actionTasksBrowserOpen() const
{
    QString agentId = tableWidget->item( tableWidget->currentRow(), ColumnAgentID )->text();

    adaptixWidget->TasksTab->SetAgentFilter(agentId);
    adaptixWidget->SetTasksUI();
}

void SessionsTableWidget::actionMarkActive() const
{
    QStringList listId;
    for( int rowIndex = 0 ; rowIndex < tableWidget->rowCount() ; rowIndex++ ) {
        if ( tableWidget->item(rowIndex, 0)->isSelected() ) {
            auto agentId = tableWidget->item( rowIndex, ColumnAgentID )->text();
            listId.append(agentId);
        }
    }

    if(listId.empty())
        return;

    QString message = QString();
    bool ok = false;
    bool result = HttpReqAgentSetMark(listId, "", *(adaptixWidget->GetProfile()), &message, &ok);
    if( !result ) {
        MessageError("Response timeout");
        return;
    }
}

void SessionsTableWidget::actionMarkInactive() const
{
    QStringList listId;
    for( int rowIndex = 0 ; rowIndex < tableWidget->rowCount() ; rowIndex++ ) {
        if ( tableWidget->item(rowIndex, 0)->isSelected() ) {
            auto agentId = tableWidget->item( rowIndex, ColumnAgentID )->text();
            listId.append(agentId);
        }
    }

    if(listId.empty())
        return;

    QString message = QString();
    bool ok = false;
    bool result = HttpReqAgentSetMark(listId, "Inactive", *(adaptixWidget->GetProfile()), &message, &ok);
    if( !result ) {
        MessageError("Response timeout");
        return;
    }
}

void SessionsTableWidget::actionItemColor() const
{
    QStringList listId;
    for( int rowIndex = 0 ; rowIndex < tableWidget->rowCount() ; rowIndex++ ) {
        if ( tableWidget->item(rowIndex, 0)->isSelected() ) {
            auto agentId = tableWidget->item( rowIndex, ColumnAgentID )->text();
            listId.append(agentId);
        }
    }

    if(listId.empty())
        return;

    QColor itemColor = QColorDialog::getColor(Qt::white, nullptr, "Select items color");
    if (itemColor.isValid()) {
        QString itemColorHex = itemColor.name();
        QString message = QString();
        bool ok = false;
        bool result = HttpReqAgentSetColor(listId, itemColorHex, "", false, *(adaptixWidget->GetProfile()), &message, &ok);
        if( !result ) {
            MessageError("Response timeout");
            return;
        }
    }
}

void SessionsTableWidget::actionTextColor() const
{
    QStringList listId;
    for( int rowIndex = 0 ; rowIndex < tableWidget->rowCount() ; rowIndex++ ) {
        if ( tableWidget->item(rowIndex, 0)->isSelected() ) {
            auto agentId = tableWidget->item( rowIndex, ColumnAgentID )->text();
            listId.append(agentId);
        }
    }

    if(listId.empty())
        return;

    QColor textColor = QColorDialog::getColor(Qt::white, nullptr, "Select text color");
    if (textColor.isValid()) {
        QString textColorHex = textColor.name();
        QString message = QString();
        bool ok = false;
        bool result = HttpReqAgentSetColor(listId, "",  textColorHex, false, *(adaptixWidget->GetProfile()), &message, &ok);
        if( !result ) {
            MessageError("Response timeout");
            return;
        }
    }
}

void SessionsTableWidget::actionColorReset() const
{
    QStringList listId;
    for( int rowIndex = 0 ; rowIndex < tableWidget->rowCount() ; rowIndex++ ) {
        if ( tableWidget->item(rowIndex, 0)->isSelected() ) {
            auto agentId = tableWidget->item( rowIndex, ColumnAgentID )->text();
            listId.append(agentId);
        }
    }

    if(listId.empty())
        return;

    QString message = QString();
    bool ok = false;
    bool result = HttpReqAgentSetColor(listId, "",  "", true, *(adaptixWidget->GetProfile()), &message, &ok);
    if( !result ) {
        MessageError("Response timeout");
        return;
    }
}

void SessionsTableWidget::actionConsoleDelete()
{
    QMessageBox::StandardButton reply = QMessageBox::question(this, "Clear Confirmation",
                                      "Are you sure you want to delete all agent console data and history from server (tasks will not be deleted from TaskManager)?\n\n"
                                      "If you want to temporarily hide the contents of the agent console, do so through the agent console menu.",
                                      QMessageBox::Yes | QMessageBox::No,
                                      QMessageBox::No);
    if (reply != QMessageBox::Yes)
        return;

    QStringList listId;
    for( int rowIndex = 0 ; rowIndex < tableWidget->rowCount() ; rowIndex++ ) {
        if ( tableWidget->item(rowIndex, 0)->isSelected() ) {
            auto agentId = tableWidget->item( rowIndex, ColumnAgentID )->text();
            listId.append(agentId);
        }
    }

    if(listId.empty())
        return;

    for (auto id : listId)
        adaptixWidget->AgentsMap[id]->Console->Clear();

    QString message = QString();
    bool ok = false;
    bool result = HttpReqConsoleRemove(listId, *(adaptixWidget->GetProfile()), &message, &ok);
    if( !result ) {
        MessageError("Response timeout");
        return;
    }
}

void SessionsTableWidget::actionAgentRemove()
{
    QMessageBox::StandardButton reply = QMessageBox::question(this, "Delete Confirmation",
                                      "Are you sure you want to delete all information about the selected agents from the server?\n\n"
                                      "If you want to hide the record, simply choose: 'Item -> Hide on Client'.",
                                      QMessageBox::Yes | QMessageBox::No,
                                      QMessageBox::No);

    if (reply != QMessageBox::Yes)
        return;

    QStringList listId;
    for( int rowIndex = 0 ; rowIndex < tableWidget->rowCount() ; rowIndex++ ) {
        if ( tableWidget->item(rowIndex, 0)->isSelected() ) {
            auto agentId = tableWidget->item( rowIndex, ColumnAgentID )->text();
            listId.append(agentId);
        }
    }

    if(listId.empty())
        return;

    QString message = QString();
    bool ok = false;
    bool result = HttpReqAgentRemove(listId, *(adaptixWidget->GetProfile()), &message, &ok);
    if( !result ) {
        MessageError("Response timeout");
        return;
    }
}

void SessionsTableWidget::actionItemTag() const
{
    QStringList listId;
    for( int rowIndex = 0 ; rowIndex < tableWidget->rowCount() ; rowIndex++ ) {
        if ( tableWidget->item(rowIndex, 0)->isSelected() ) {
            auto agentId = tableWidget->item( rowIndex, ColumnAgentID )->text();
            listId.append(agentId);
        }
    }

    if(listId.empty())
        return;

    QString tag = "";
    if(listId.size() == 1) {
        tag = tableWidget->item( tableWidget->currentRow(), ColumnTags )->text();
    }

    bool inputOk;
    QString newTag = QInputDialog::getText(nullptr, "Set tags", "New tag", QLineEdit::Normal,tag, &inputOk);
    if ( inputOk ) {
        QString message = QString();
        bool ok = false;
        bool result = HttpReqAgentSetTag(listId, newTag, *(adaptixWidget->GetProfile()), &message, &ok);
        if( !result ) {
            MessageError("Response timeout");
            return;
        }
    }
}

void SessionsTableWidget::actionItemHide() const
{
    for( int rowIndex = 0 ; rowIndex < tableWidget->rowCount() ; rowIndex++ ) {
        if ( tableWidget->item(rowIndex, 0)->isSelected() ) {

            auto agentId = tableWidget->item( rowIndex, ColumnAgentID )->text();

            if (adaptixWidget->AgentsMap.contains(agentId))
                adaptixWidget->AgentsMap[agentId]->show = false;

        }
    }

    this->SetData();
}

void SessionsTableWidget::actionItemsShowAll() const
{
    bool refact = false;
    for (auto agent : adaptixWidget->AgentsMap) {
        if (agent->show == false) {
            agent->show = true;
            refact = true;
        }
    }

    if (refact)
        this->SetData();
}