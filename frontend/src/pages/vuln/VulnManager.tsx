import { Tabs, TabsContent, TabsList, TabsTrigger } from "../../components/ui/Tabs";
import VulnList from "./VulnList";
import { PythonRunner } from "../../components/poc/PythonRunner";
import { NucleiRunner } from "../../components/poc/NucleiRunner";

export default function VulnManager() {
    return (
        <div className="h-full p-4">
            <Tabs defaultValue="info" className="h-full flex flex-col">
                <div className="flex justify-center mb-4">
                    <TabsList className="grid w-full grid-cols-3 max-w-[600px]">
                        <TabsTrigger value="info">信息管理</TabsTrigger>
                        <TabsTrigger value="python">Python 验证</TabsTrigger>
                        <TabsTrigger value="nuclei">Nuclei 验证</TabsTrigger>
                    </TabsList>
                </div>

                <TabsContent value="info" className="flex-1 mt-0 overflow-hidden">
                    <VulnList />
                </TabsContent>
                
                <TabsContent value="python" className="flex-1 mt-0 overflow-hidden">
                    <PythonRunner />
                </TabsContent>

                <TabsContent value="nuclei" className="flex-1 mt-0 overflow-hidden">
                    <NucleiRunner />
                </TabsContent>
            </Tabs>
        </div>
    );
}
